package whitenoise

import (
	"bufio"
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
	peerMap       PeerMapper
	sessionMapper SessionMapper
}

type PeerMapper struct {
	peerMap  map[core.PeerID]core.PeerAddrInfo
	peerChan chan core.PeerAddrInfo
	event    chan core.PeerAddrInfo
}

func (self *PeerMapper) run() {
	for {
		peer := <-self.peerChan
		println("get peer:", peer.ID)
		_, ok := self.peerMap[peer.ID]
		if !ok {
			self.event <- peer
		}
		self.peerMap[peer.ID] = peer
	}
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
	host.SetStreamHandler(protocol.ID(WHITENOISE_PROTOCOL), StreamHandler)
	service := NetworkService{
		host: host,
		ctx:  ctx,
		peerMap: PeerMapper{
			peerMap:  make(map[core.PeerID]core.PeerAddrInfo),
			peerChan: peerChan,
			event:    make(chan core.PeerAddrInfo),
		},
	}
	return &service, nil

}

func (service *NetworkService) Start() {
	println("start service")
	go service.peerMap.run()
	for {
		peer := <-service.peerMap.event
		service.TryConnect(peer)
		service.GenWhiteNoiseStream(peer.ID)
	}
}

func (service *NetworkService) GenWhiteNoiseStream(peerID core.PeerID) {
	stream, err := service.host.NewStream(service.ctx, peerID, protocol.ID(WHITENOISE_PROTOCOL))

	if err != nil {
		fmt.Printf("newstream to %v error: %v\n", peerID, err)
	} else {
		fmt.Println("gen new stream: ", stream.ID())
		rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))
		go writeData(rw)
		go readData(rw)
		//session := NewSession(stream)
		//service.sessionMapper.AddSession(session)
		fmt.Printf("Connected to:%v \n", peerID)
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
