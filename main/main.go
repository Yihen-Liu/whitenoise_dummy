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
	if cfg.Proxy != "" {
		time.Sleep(time.Second)
		proxy, err := peer.Decode(cfg.Proxy)
		if err != nil {
			log.Error(err)
		}
		err = service.RegisterProxy(proxy)
		if err != nil {
			log.Error(err)
		}
	}

	if cfg.Session != "" {
		time.Sleep(time.Second * 2)
		id, err := peer.Decode(cfg.Session)
		if err != nil {
			log.Error(err)
		}
		err = service.NewSessionToPeer(id, "hello")
		if err != nil {
			log.Error(err)
		}
	}

	//if cfg.Expend != "" && cfg.Session != "" {
	//	time.Sleep(time.Second * 2)
	//	to, err := peer.Decode(cfg.Session)
	//	if err != nil {
	//		panic(err)
	//	}
	//	expend, err := peer.Decode(cfg.Expend)
	//	if err != nil {
	//		panic(err)
	//	}
	//	err = service.ExpendSession(to, expend, "hello")
	//	if err != nil {
	//		panic(err)
	//	}
	//}

	//if cfg.Des != "" && cfg.Join != "" {
	//	time.Sleep(time.Second * 2)
	//	join, err := peer.Decode(cfg.Join)
	//	if err != nil {
	//		panic(err)
	//	}
	//	des, err := peer.Decode(cfg.Des)
	//	if err != nil {
	//		panic(err)
	//	}
	//	err = service.NewSessionToPeer(join, "hello")
	//	if err != nil {
	//		panic(err)
	//	}
	//	err = service.GossipJoint(des, join, "hello")
	//	if err != nil {
	//		panic(err)
	//	}
	//}

	if cfg.Session != "" && cfg.Circuit != "" {
		time.Sleep(time.Second * 2)
		des, err := peer.Decode(cfg.Circuit)
		if err != nil {
			panic(err)
		}
		err = service.NewCircuit(des,"hello")
		if err != nil {
			panic(err)
		}
	}

	if cfg.Relay {
		time.Sleep(time.Second * 5)
		service.SendRelay("hello", []byte("HoHoHo!"))
	}

	//if cfg.Pub != "" {
	//	time.Sleep(time.Second * 2)
	//	err := service.PubsubService.Publish([]byte(cfg.Pub))
	//	if err != nil {
	//		log.Error("Publish err", err)
	//	}
	//}

	time.Sleep(time.Second * 1000)
}
