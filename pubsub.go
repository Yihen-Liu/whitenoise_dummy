package whitenoise

import (
	"context"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"whitenoise/log"
)

const GOSSIPTOPICNAME string = "gossip_topic"

type PubsubService struct {
	// Messages is a channel of messages received from other peers in the chat room
	Messages chan GossipMsg
	ctx      context.Context
	ps       *pubsub.PubSub
	topic    *pubsub.Topic
	sub      *pubsub.Subscription
}

type GossipMsg struct {
	data []byte
}

func (service *NetworkService) NewPubsubService() error {
	ctx, _ := context.WithCancel(service.ctx)
	ps, err := pubsub.NewGossipSub(ctx, service.host)
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
	//msgBytes, err := json.Marshal(data)
	//if err != nil {
	//	return err
	//}
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
		//err = json.Unmarshal(msg.Data, &gossipMsg)
		//if err != nil {
		//	continue
		//}
		service.Messages <- gossipMsg
	}
}
