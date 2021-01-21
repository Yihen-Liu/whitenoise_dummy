package main

import (
	"context"
	"github.com/libp2p/go-libp2p-core/peer"
	"time"
	"whitenoise/log"
)
import "whitenoise"

func main() {
	ctx := context.Background()
	cfg := whitenoise.NewConfig()
	host, err := whitenoise.NewDummyHost(ctx, cfg)
	if err != nil {
		panic(err)
	}
	service, err := whitenoise.NewService(ctx, host, cfg)
	if err != nil {
		panic(err)
	}
	service.Start()
	///for testing
	if cfg.Session != "" {
		time.Sleep(time.Second * 2)
		//i := 0
		//for _, v := range service.SessionMapper.StreamMap {
		//	println("start set session id")
		//	service.SetSessionId("hello"+string(rune(i)), v.StreamId)
		//	i++
		//}
		id, err := peer.Decode(cfg.Session)
		if err != nil {
			println(err)
		}
		err = service.NewSessionToPeer(id, "hello")
		if err != nil {
			println(err)
		}
	}

	if cfg.Relay {
		time.Sleep(time.Second * 2)
		//for _, session := range service.SessionMapper.SessionmapID {
		//	service.SendRelay(session.Id, []byte("Hi..Hi..Hi.."))
		//}
		service.SendRelay("hello", []byte("HoHoHo!"))
	}

	if cfg.Pub != "" {
		time.Sleep(time.Second * 2)
		err := service.PubsubService.Publish([]byte(cfg.Pub))
		if err != nil {
			log.Error("Publish err", err)
		}
	}

	time.Sleep(time.Second * 1000)
}
