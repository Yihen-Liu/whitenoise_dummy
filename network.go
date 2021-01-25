package whitenoise

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/libp2p/go-libp2p"
	core "github.com/libp2p/go-libp2p-core"
	crypto "github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/protocol"
	noise "github.com/libp2p/go-libp2p-noise"
	"github.com/multiformats/go-multiaddr"
	"time"
	"whitenoise/log"
	"whitenoise/pb"
)

const SETSESSIONTIMEOUT time.Duration = time.Second
const EXPENDSESSIONTIMEOUT time.Duration = time.Second * 3

type NetworkService struct {
	host          host.Host
	ctx           context.Context
	discovery     Discovery
	SessionMapper SessionManager
	inboundEvent  chan InboundEvent
	PubsubService *PubsubService
	AckManager    *AckManager
}

func (service *NetworkService) TryConnect(peer core.PeerAddrInfo) {
	err := service.host.Connect(service.ctx, peer)
	if err != nil {
		log.Errorf("connect to %v error: %v", peer.ID, err)
	} else {
	}
}

func NewService(ctx context.Context, host host.Host, cfg *NetworkConfig) (*NetworkService, error) {
	peerChan := initMDNS(ctx, host, cfg.RendezvousString)

	service := NetworkService{
		host: host,
		ctx:  ctx,
		discovery: Discovery{
			PeerMap:  make(map[core.PeerID]core.PeerAddrInfo),
			peerChan: peerChan,
			event:    make(chan core.PeerAddrInfo),
		},
		SessionMapper: NewSessionMapper(),
		inboundEvent:  make(chan InboundEvent),
		AckManager:    NewAckManager(),
	}
	err := service.NewPubsubService()
	if err != nil {
		log.Error("New service err: ", err)
	}
	service.host.SetStreamHandler(protocol.ID(RELAY_PROTOCOL), service.RelayStreamHandler)
	service.host.SetStreamHandler(protocol.ID(CMD_PROTOCOL), service.CmdStreamHandler)
	service.host.SetStreamHandler(protocol.ID(ACK_PROTOCOL), service.AckManager.AckStreamHandler)
	return &service, nil
}

func (service *NetworkService) Start() {
	log.Infof("start service %v\n", peer.Encode(service.host.ID()))
	go service.discovery.run()
	service.PubsubService.Start()
	go service.PubsubService.HandleGossipMsg()
	go func() {
		for {
			peer := <-service.discovery.event
			service.TryConnect(peer)
		}
	}()
}

func (service *NetworkService) NewRelayStream(peerID core.PeerID) (string, error) {
	stream, err := service.host.NewStream(service.ctx, peerID, protocol.ID(RELAY_PROTOCOL))
	if err != nil {
		log.Infof("newstream to %v error: %v\n", peerID, err)
		return "", err
	}
	log.Info("gen new stream: ", stream.ID())
	s := NewStream(stream)
	service.SessionMapper.AddSessionNonid(s)
	go service.RelayInboundHandler(s)
	log.Infof("Connected to:%v \n", peerID)
	_, err = s.RW.Write(NewMsg([]byte("hi")))
	if err != nil {
		log.Error("write err", err)
		return "", err
	}
	err = s.RW.Flush()
	if err != nil {
		log.Error("flush err", err)
		return "", err
	}
	return s.StreamId, nil
}

func (service *NetworkService) NewSessionToPeer(peerID core.PeerID, sessionID string) error {
	streamId, err := service.NewRelayStream(peerID)
	if err != nil {
		return err
	}
	err = service.SetSessionId(sessionID, streamId)
	return err
}

func (service *NetworkService) SetSessionId(sessionID string, streamID string) error {
	stream, ok := service.SessionMapper.StreamMap[streamID]
	if ok {
		s, ok := service.SessionMapper.SessionmapID[sessionID]
		if !ok {
			s = NewSession()
			s.SetSessionID(sessionID)
		}
		s.AddStream(stream)
		service.SessionMapper.AddSessionId(sessionID, s)
		log.Infof("session: %v\n", s)
	} else {
		return errors.New("no such stream:" + streamID)
	}

	data, id := NewSetSessionIDCommand(sessionID, streamID)
	_, err := stream.RW.Write(data)
	if err != nil {
		log.Error("write err", err)
		return err
	}
	err = stream.RW.Flush()
	if err != nil {
		log.Error("flush err", err)
		return err
	}
	res := Task{
		Id:      id,
		channel: make(chan bool),
	}
	service.AckManager.TaskMap[id] = res

	defer delete(service.AckManager.TaskMap, id)
	select {
	case <-time.After(SETSESSIONTIMEOUT):
		err := errors.New("timeout")
		return err
	case ok := <-res.channel:
		if !ok {
			return errors.New("cmd rejected")
		}
	}
	return nil
}

func (service *NetworkService) SendRelay(sessionid string, data []byte) {
	session, ok := service.SessionMapper.SessionmapID[sessionid]
	if !ok {
		log.Info("SendRelay no such session")
	}
	payload := NewRelay(data, sessionid)
	for _, stream := range session.GetPair() {
		_, err := stream.RW.Write(payload)
		if err != nil {
			log.Error("write err", err)
			return
		}
		err = stream.RW.Flush()
		if err != nil {
			log.Error("flush err", err)
			return
		}
	}

}

func (service *NetworkService) ExpendSession(to core.PeerID, expend core.PeerID, sessionId string) error {
	stream, err := service.host.NewStream(service.ctx, to, protocol.ID(CMD_PROTOCOL))
	if err != nil {
		return err
	}
	s := NewStream(stream)
	cmd := pb.SessionExpend{
		SessionId: sessionId,
		PeerId:    expend.String(),
	}
	cmdData, err := proto.Marshal(&cmd)
	if err != nil {
		return err
	}
	payload := pb.Command{
		CommandId: "",
		Type:      pb.Cmdtype_SessionExPend,
		From:      service.host.ID().String(),
		Data:      cmdData,
	}
	payloadNoId, err := proto.Marshal(&payload)
	if err != nil {
		return err
	}
	hash := sha256.Sum256(payloadNoId)
	payload.CommandId = EncodeID(hash[:])

	data, err := proto.Marshal(&payload)
	if err != nil {
		return err
	}

	encode := EncodePayload(data)
	_, err = s.RW.Write(encode)
	if err != nil {
		return err
	}
	err = s.RW.Flush()
	if err != nil {
		return err
	}
	task := Task{
		Id:      payload.CommandId,
		channel: make(chan bool),
	}
	service.AckManager.TaskMap[payload.CommandId] = task
	defer delete(service.AckManager.TaskMap, payload.CommandId)
	select {
	case <-time.After(EXPENDSESSIONTIMEOUT):
		return errors.New("timeout")
	case ok := <-task.channel:
		if !ok {
			return errors.New("cmd rejected")
		}
	}
	return nil
}

func (service *NetworkService) GossipJoint(des peer.ID, join peer.ID, sessionId string) error {
	var neg = pb.Negotiate{
		Id:          "",
		Join:        join.String(),
		SessionId:   sessionId,
		Destination: des.String(),
		Sig:         []byte{},
	}

	negNoID, err := proto.Marshal(&neg)
	if err != nil {
		return err
	}
	hash := sha256.Sum256(negNoID)
	neg.Id = EncodeID(hash[:])
	data, err := proto.Marshal(&neg)
	if err != nil {
		return err
	}

	err = service.PubsubService.Publish(data)
	if err != nil {
		return err
	}
	return nil
}

func NewDummyHost(ctx context.Context, cfg *NetworkConfig) (host.Host, error) {
	r := rand.Reader
	priv, _, err := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, r)
	if err != nil {
		return nil, err
	}
	sourceMultiAddr, _ := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/%s/tcp/%d", cfg.listenHost, cfg.listenPort))
	transport, err := noise.New(priv)
	if err != nil {
		return nil, err
	}
	host, err := libp2p.New(
		ctx,
		libp2p.ListenAddrs(sourceMultiAddr),
		libp2p.Security(noise.ID, transport),
		libp2p.Identity(priv))
	if err != nil {
		return nil, err
	}
	return host, nil
}
