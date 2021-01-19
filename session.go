package whitenoise

import (
	"bufio"
	"github.com/libp2p/go-libp2p-core/network"
)

type Session struct {
	Id       string
	StreamId string
	rw       *bufio.ReadWriter
}

func NewSession(stream network.Stream) Session {
	return Session{
		Id:       stream.ID(),
		StreamId: stream.ID(),
		rw:       bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream)),
	}
}

type SessionMapper struct {
	SessionMap map[string]Session
}

func (mapper SessionMapper) AddSession(s Session) {
	mapper.SessionMap[s.Id] = s
}
