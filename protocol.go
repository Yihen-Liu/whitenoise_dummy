package whitenoise

import (
	"bufio"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/libp2p/go-libp2p-core/network"
	"io"
	"os"
	"whitenoise/pb"
)

const WHITENOISE_PROTOCOL string = "whitenoise"

type InboundEvent struct {
	//raw       []byte
	//sessionID string
	//streamID  string
	//data      []byte
	msg []byte
}

func (s *NetworkService) StreamHandler(stream network.Stream) {
	fmt.Println("Got a new stream: ", stream.ID())
	str := NewStream(stream)
	s.SessionMapper.AddSessionNonid(str)
	go s.InboundHandler(str)
}

func (service *NetworkService) InboundHandler(s Stream) {
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
			println("payload not enough bytes")
			continue
		}

		var payload pb.Payload = pb.Payload{}
		err = proto.Unmarshal(msgBytes, &payload)
		if err != nil {
			println("unmarshal err", err)
			continue
		}
		//TODO:add command dispatch
		if !payload.SessionCmd {
			if payload.SessionId == "" {
				service.handleMsg(&payload, s)
				continue
			} else {
				//HandleRelay
				println("Handling relay")
				session, ok := service.SessionMapper.SessionmapID[payload.SessionId]
				if ok {
					if session.IsReady() {
						part, err := session.GetPattern(s.StreamId)
						if err != nil {
							println("get part err", err)
							continue
						}
						relay := NewRelay(payload.Data, payload.SessionId)
						_, err = part.RW.Write(relay)
						if err != nil {
							println("write err", err)
							continue
						}
						err = part.RW.Flush()
						if err != nil {
							println("flush err", err)
							continue
						}
					} else {
						//TODO:give relay msg to client for relay node
						fmt.Printf("maybe endpoint %v\n", string(payload.Data))
					}
				} else {
					println("relay no such session")
				}
			}

		} else {
			err = service.handleCommand(&payload, s)
			if err != nil {
				println("handle command err: ", err)
				continue
			}
		}
	}
}

func (service NetworkService) handleMsg(payload *pb.Payload, s Stream) {
	fmt.Printf("Receive msg:%v\n", string(payload.Data))
}

func (service NetworkService) handleCommand(payload *pb.Payload, s Stream) error {
	session, ok := service.SessionMapper.SessionmapID[payload.SessionId]
	if !ok {
		session = NewSession()
		session.SetSessionID(payload.SessionId)
	}
	session.AddStream(s)
	service.SessionMapper.AddSessionId(payload.SessionId, session)
	fmt.Printf("add sessionid %v to stream %v\n", payload.SessionId, s.StreamId)
	fmt.Printf("session: %v\n", session)

	reply := NewMsg([]byte("Reply add sessionid " + payload.SessionId + " to stream" + s.StreamId))
	_, err := s.RW.Write(reply)
	if err != nil {
		println("write err", err)
		return err
	}
	err = s.RW.Flush()
	if err != nil {
		println("flush err", err)
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
