package whitenoise

import (
	"flag"
)

//type NetworkConfig struct {
//}
//
//func NewConfig() {
//
//}

type NetworkConfig struct {
	RendezvousString string
	ProtocolID       string
	listenHost       string
	listenPort       int
	Session          bool
}

func NewConfig() *NetworkConfig {
	return parseFlags()
}

func parseFlags() *NetworkConfig {
	c := &NetworkConfig{}

	flag.StringVar(&c.RendezvousString, "rendezvous", "meetme", "Unique string to identify group of nodes. Share this with your friends to let them connect with you")
	flag.StringVar(&c.listenHost, "host", "127.0.0.1", "The bootstrap node host listen address\n")
	flag.StringVar(&c.ProtocolID, "pid", "whitenoise", "Sets a protocol id for stream headers")
	flag.IntVar(&c.listenPort, "port", 4001, "node listen port")
	flag.BoolVar(&c.Session,"session",false,"test session")
	flag.Parse()
	return c
}
