package whitenoise

import (
	"github.com/libp2p/go-libp2p-core/network"
	"io"
	"whitenoise/log"
)

const CMDPROTOCOL string = "cmd"

type CmdHandler struct {
	service *NetworkService
	CmdList map[string]bool
}

func (c *CmdHandler) run() {

}

func (s *NetworkService) CmdStreamHandler(stream network.Stream) {
	defer stream.Close()
	str := NewStream(stream)
	lBytes := make([]byte, 4)
	_, err := io.ReadFull(str.RW, lBytes)
	if err != nil {
		return
	}

	l := Bytes2Int(lBytes)
	msgBytes := make([]byte, l)
	_, err = io.ReadFull(str.RW, msgBytes)
	if err != nil {
		log.Warn("payload not enough bytes")
		return
	}




}

