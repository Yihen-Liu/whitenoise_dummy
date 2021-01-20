package main

import (
	"context"
	"github.com/libp2p/go-libp2p-core/peer"
	"time"
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
		time.Sleep(time.Second * 5)
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
		time.Sleep(time.Second * 5)
		//for _, session := range service.SessionMapper.SessionmapID {
		//	service.SendRelay(session.Id, []byte("Hi..Hi..Hi.."))
		//}
		service.SendRelay("hello",[]byte("HoHoHo!"))
	}

	time.Sleep(time.Second * 1000)
}
