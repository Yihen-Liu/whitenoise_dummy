package whitenoise

import (
	"context"
	"crypto/rand"
	"fmt"
	"github.com/libp2p/go-libp2p"
	core "github.com/libp2p/go-libp2p-core"
	crypto "github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/protocol"
	noise "github.com/libp2p/go-libp2p-noise"
	"github.com/multiformats/go-multiaddr"
)

type NetworkService struct {
	host          host.Host
	ctx           context.Context
	peerMap       Discovery
	SessionMapper SessionMapper
	inboundEvent  chan InboundEvent
}

func (service *NetworkService) TryConnect(peer core.PeerAddrInfo) {
	err := service.host.Connect(service.ctx, peer)
	if err != nil {
		fmt.Printf("connect to %v error: %v", peer.ID, err)
	} else {
	}
}

func NewService(ctx context.Context, host host.Host, cfg *NetworkConfig) (*NetworkService, error) {
	peerChan := initMDNS(ctx, host, cfg.RendezvousString)

	service := NetworkService{
		host: host,
		ctx:  ctx,
		peerMap: Discovery{
			peerMap:  make(map[core.PeerID]core.PeerAddrInfo),
			peerChan: peerChan,
			event:    make(chan core.PeerAddrInfo),
		},
		SessionMapper: NewSessionMapper(),
		inboundEvent:  make(chan InboundEvent),
	}
	service.host.SetStreamHandler(protocol.ID(WHITENOISE_PROTOCOL), service.StreamHandler)
	return &service, nil
}

func (service *NetworkService) Start() {
	println("start service")
	go service.peerMap.run()
	go func() {
		for {
			peer := <-service.peerMap.event
			service.TryConnect(peer)
			service.NewWhiteNoiseStream(peer.ID)
		}
	}()
}

func (service *NetworkService) NewWhiteNoiseStream(peerID core.PeerID) {
	stream, err := service.host.NewStream(service.ctx, peerID, protocol.ID(WHITENOISE_PROTOCOL))
	if err != nil {
		fmt.Printf("newstream to %v error: %v\n", peerID, err)
	} else {
		fmt.Println("gen new stream: ", stream.ID())
		s := NewSession(stream)
		service.SessionMapper.AddSessionNonid(s)
		go service.InboundHandler(s)
		fmt.Printf("Connected to:%v \n", peerID)
		_, err = s.RW.Write(NewMsg([]byte("hi")))
		if err != nil {
			println("write err", err)
			return
		}
		err = s.RW.Flush()
		if err != nil {
			println("flush err", err)
			return
		}
	}
}

func (service *NetworkService) SetSessionId(sessionID string, streamID string) {
	s, ok := service.SessionMapper.SessionmapNonid[streamID]
	if ok {
		service.SessionMapper.AddSessionId(sessionID, s)
	} else {
		println("no such stream:", streamID)
	}
	_, err := s.RW.Write(NewSetSessionIDCommand(sessionID, streamID))
	if err != nil {
		println("write err", err)
		return
	}
	err = s.RW.Flush()
	if err != nil {
		println("flush err", err)
		return
	}
}

func (service *NetworkService) ConnectSession() {

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
