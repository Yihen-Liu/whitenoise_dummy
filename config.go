package whitenoise

import (
	"flag"
)

type NetworkConfig struct {
	RendezvousString string
	ProtocolID       string
	listenHost       string
	listenPort       int
	Session          string
	Relay            bool
	Pub              string
	Expend           string
	Des              string
	Join             string
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
	flag.StringVar(&c.Session, "session", "", "test session peerId")
	flag.StringVar(&c.Pub, "pub", "", "test gossip pubsub")
	flag.BoolVar(&c.Relay, "relay", false, "test relay")
	flag.StringVar(&c.Expend, "expend", "", "test session expend")
	flag.StringVar(&c.Des, "des", "", "gossip destination peerid")
	flag.StringVar(&c.Join, "join", "", "gossip join node peerid")
	flag.Parse()
	return c
}
