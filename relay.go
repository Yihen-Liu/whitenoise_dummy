package whitenoise

import (
	"bufio"
	"fmt"
	"github.com/golang/protobuf/proto"
	core "github.com/libp2p/go-libp2p-core"
	"github.com/libp2p/go-libp2p-core/network"
	"io"
	"os"
	"whitenoise/log"
	"whitenoise/pb"
)

const RELAY_PROTOCOL string = "relay"

type InboundEvent struct {
	//raw       []byte
	//sessionID string
	//streamID  string
	//data      []byte
}

func (s *NetworkService) RelayStreamHandler(stream network.Stream) {
	fmt.Println("Got a new stream: ", stream.ID())
	str := NewStream(stream)
	s.SessionMapper.AddSessionNonid(str)
	go s.RelayInboundHandler(str)
}

func (service *NetworkService) RelayInboundHandler(s Stream) {
	for {
		lBytes := make([]byte, 4)
		_, err := io.ReadFull(s.RW, lBytes)
		if err != nil {
			continue
		}

		l := Bytes2Int(lBytes)
		msgBytes := make([]byte, l)
		_, err = io.ReadFull(s.RW, msgBytes)
		if err != nil {
			log.Info("payload not enough bytes")
			continue
		}

		var payload pb.Payload = pb.Payload{}
		err = proto.Unmarshal(msgBytes, &payload)
		if err != nil {
			log.Error("unmarshal err", err)
			continue
		}

		//TODO:add command dispatch
		if !payload.SessionCmd {
			if payload.SessionId == "" {
				service.handleMsg(&payload, s)
				continue
			} else {
				//HandleRelay
				log.Info("Handling relay")
				session, ok := service.SessionMapper.SessionmapID[payload.SessionId]
				if ok {
					if session.IsReady() {
						part, err := session.GetPattern(s.StreamId)
						if err != nil {
							log.Error("get part err", err)
							continue
						}
						relay := NewRelay(payload.Data, payload.SessionId)
						_, err = part.RW.Write(relay)
						if err != nil {
							log.Error("write err", err)
							continue
						}
						err = part.RW.Flush()
						if err != nil {
							log.Error("flush err", err)
							continue
						}
					} else {
						//TODO:give relay msg to client for relay node
						log.Infof("maybe endpoint %v\n", string(payload.Data))
					}
				} else {
					log.Warn("relay no such session")
				}
			}

		} else {
			err = service.handleSetSession(&payload, s)
			if err != nil {
				log.Error("handle command err: ", err)
				continue
			}
		}

	}
}

func (service *NetworkService) handleMsg(payload *pb.Payload, s Stream) {
	log.Infof("Receive msg:%v\n", string(payload.Data))
}

func (service *NetworkService) handleSetSession(payload *pb.Payload, s Stream) error {
	session, ok := service.SessionMapper.SessionmapID[payload.SessionId]
	if !ok {
		session = NewSession()
		session.SetSessionID(payload.SessionId)
	}
	session.AddStream(s)
	service.SessionMapper.AddSessionId(payload.SessionId, session)
	log.Infof("add sessionid %v to stream %v\n", payload.SessionId, s.StreamId)
	log.Infof("session: %v\n", session)

	//reply := NewMsg([]byte("Reply add sessionid " + payload.SessionId + " to stream" + s.StreamId))
	//ack
	stream, err := service.host.NewStream(service.ctx, s.RemotePeer, core.ProtocolID(ACK_PROTOCOL))
	if err != nil {
		log.Errorf("handleSetSession new stream err %v", err)
		return err
	}
	resStream := NewStream(stream)
	var ack = pb.Ack{
		CommandId: payload.Id,
		Result:    true,
		Data:      []byte{},
	}

	data, err := proto.Marshal(&ack)
	if err != nil {
		return err
	}

	encoded := EncodePayload(data)

	_, err = resStream.RW.Write(encoded)
	if err != nil {
		log.Error("write err", err)
		return err
	}
	err = resStream.RW.Flush()
	if err != nil {
		log.Error("flush err", err)
		return err
	}
	return nil
}

func readData(rw *bufio.ReadWriter) {
	for {
		str, err := rw.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading from buffer")
			panic(err)
		}

		if str == "" {
			return
		}
		if str != "\n" {
			// Green console colour: 	\x1b[32m
			// Reset console colour: 	\x1b[0m
			fmt.Printf("\x1b[32m%s\x1b[0m> ", str)
		}

	}
}

func writeData(rw *bufio.ReadWriter) {
	stdReader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("> ")
		sendData, err := stdReader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading from stdin")
			panic(err)
		}

		_, err = rw.WriteString(fmt.Sprintf("%s\n", sendData))
		if err != nil {
			fmt.Println("Error writing to buffer")
			panic(err)
		}
		err = rw.Flush()
		if err != nil {
			fmt.Println("Error flushing buffer")
			panic(err)
		}
	}
}
