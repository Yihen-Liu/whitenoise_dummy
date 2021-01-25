package whitenoise

import (
	"github.com/golang/protobuf/proto"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"io"
	"whitenoise/log"
	"whitenoise/pb"
)

const CMD_PROTOCOL string = "cmd"

type CmdHandler struct {
	service *NetworkService
	CmdList map[string]bool
}

func (c *CmdHandler) run() {
}

//change to cmdHandler
func (s *NetworkService) CmdStreamHandler(stream network.Stream) {
	defer stream.Close()
	str := NewStream(stream)
	lBytes := make([]byte, 4)
	_, err := io.ReadFull(str.RW, lBytes)
	if err != nil {
		return
	}

	l := Bytes2Int(lBytes)
	payloadBytes := make([]byte, l)
	_, err = io.ReadFull(str.RW, payloadBytes)
	if err != nil {
		log.Warn("payload not enough bytes")
		return
	}

	//dispatch
	var payload = pb.Command{}
	err = proto.Unmarshal(payloadBytes, &payload)
	if err != nil {
		log.Error("unmarshal err", err)

	}
	switch payload.Type {
	case pb.Cmdtype_SessionExPend:
		log.Info("Get session expend cmd")
		var cmd = pb.SessionExpend{}
		err = proto.Unmarshal(payload.Data, &cmd)
		if err != nil {

		}
		session, ok := s.SessionMapper.SessionmapID[cmd.SessionId]
		if !ok {
			log.Warnf("No such session: %v", cmd.SessionId)
			break
		}
		if session.IsReady() {
			log.Warnf("Session is ready, cant expend %v", cmd.SessionId)
			break
		}
		id, err := peer.Decode(cmd.PeerId)
		if err != nil {
			log.Warnf("Decode peerid %v err: %v", cmd.PeerId, err)
		}
		err = s.NewSessionToPeer(id, cmd.SessionId)
		if err != nil {
			log.Errorf("NewSessionToPeer err: %v", err)
		}
	case pb.Cmdtype_Disconnect:
		log.Info("get Disconnect cmd")
	}

}
