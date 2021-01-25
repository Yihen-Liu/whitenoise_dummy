package whitenoise

import (
	"context"
	"github.com/golang/protobuf/proto"
	core "github.com/libp2p/go-libp2p-core"
	"github.com/libp2p/go-libp2p-core/peer"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"time"
	"whitenoise/log"
	"whitenoise/pb"
)

const GOSSIPTOPICNAME string = "gossip_topic"

type PubsubService struct {
	// Messages is a channel of messages received from other peers in the chat room
	Messages chan GossipMsg
	ctx      context.Context
	ps       *pubsub.PubSub
	topic    *pubsub.Topic
	sub      *pubsub.Subscription
	network  *NetworkService
}

type GossipMsg struct {
	data []byte
}

func (service *NetworkService) NewPubsubService() error {
	ctx, _ := context.WithCancel(service.ctx)
	ps, err := pubsub.NewGossipSub(ctx, service.host, pubsub.WithNoAuthor()) //pubsub omit from,dig and change default id
	if err != nil {
		log.Error("NewPubsubService err: ", err)
		return err
	}
	topic, err := ps.Join(GOSSIPTOPICNAME)
	if err != nil {
		log.Error("NewPubsubService err: ", err)
		return err
	}
	sub, err := topic.Subscribe()
	if err != nil {
		log.Error("NewPubsubService err: ", err)
		return err
	}

	pubsubService := PubsubService{
		Messages: make(chan GossipMsg),
		ctx:      ctx,
		ps:       ps,
		topic:    topic,
		sub:      sub,
		network:  service,
	}
	service.PubsubService = &pubsubService
	return nil
}

func (service *PubsubService) Start() {
	go service.handleMessage()
}

func (service *PubsubService) Subscribe() {

}

func (service *PubsubService) Publish(data []byte) error {
	return service.topic.Publish(service.ctx, data)
}

func (service *PubsubService) handleMessage() {
	for {
		msg, err := service.sub.Next(service.ctx)
		if err != nil {
			close(service.Messages)
			return
		}

		gossipMsg := GossipMsg{data: msg.Data}
		service.Messages <- gossipMsg
	}
}

func (service *PubsubService) HandleGossipMsg() {
	for {
		msg := <-service.Messages
		log.Infof("Receive Gossip: %v\n", string(msg.data))
		var neg = pb.Negotiate{}
		err := proto.Unmarshal(msg.data, &neg)
		if err != nil {
			log.Errorf("Unmarshall gossip error: %v", err)
			continue
		}
		destination, err := peer.Decode(neg.Destination)
		if err != nil {
			log.Errorf("Decode destination error: %v", err)
			continue
		}
		if destination != service.network.host.ID() {
			log.Info("Gossip not for me")
			continue
		}
		_, ok := service.network.SessionMapper.SessionmapID[neg.SessionId]
		if ok {
			log.Errorf("session already exist %v", neg.SessionId)
		}
		joinNode, err := peer.Decode(neg.Join)
		var relayId core.PeerID
		for id, _ := range service.network.discovery.PeerMap {
			//todo:筛选中继节点，避免入口节点作为中继节点
			if id != joinNode {
				relayId = id
				break
			}
		}

		//todo：错误处理（重试）
		err = service.network.NewSessionToPeer(relayId, neg.SessionId)
		if err != nil {
			log.Errorf("New session to relay err %v", err)
			continue
		}

		time.Sleep(time.Second)

		err = service.network.ExpendSession(relayId, joinNode, neg.SessionId)
		if err != nil {
			log.Errorf("Expend session err %v", err)
			continue
		}

	}
}
