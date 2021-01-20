package whitenoise

import (
	"bufio"
	"github.com/libp2p/go-libp2p-core/network"
)

const SESSIONIDNON string = "SessionIDNon"

//TODO:add mutex
type Session struct {
	Id       string
	StreamId string
	RW       *bufio.ReadWriter
}

func NewSession(stream network.Stream) Session {
	return Session{
		Id:       SESSIONIDNON,
		StreamId: stream.ID(),
		RW:       bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream)),
	}
}

func (s *Session) SetSessionID(sID string) {
	s.Id = sID
}

type SessionMapper struct {
	SessionmapNonid map[string]Session
	SessionmapID    map[string]Session
}

func NewSessionMapper() SessionMapper {
	return SessionMapper{
		SessionmapNonid: make(map[string]Session),
		SessionmapID:    make(map[string]Session),
	}
}

func (mapper *SessionMapper) AddSessionNonid(s Session) {
	mapper.SessionmapNonid[s.StreamId] = s
}

func (mapper *SessionMapper) AddSessionId(id string, s Session) {
	mapper.SessionmapID[id] = s
}
