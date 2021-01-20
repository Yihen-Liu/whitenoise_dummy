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
}

func (s *NetworkService) StreamHandler(stream network.Stream) {
	fmt.Println("Got a new stream: ", stream.ID())
	session := NewSession(stream)
	s.SessionMapper.AddSessionNonid(session)
	go s.InboundHandler(session)
}

func (service *NetworkService) InboundHandler(s Session) {
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
		if payload.SessionId == "" {
			fmt.Printf("Receive msg:%v\n", string(payload.Data))
			continue
		}
		//s, ok := service.SessionMapper.SessionmapNonid[payload.StreamId]
		//if !ok {
		//	fmt.Printf("Receive set session request but no such stream %v\n", payload.StreamId)
		//	continue
		//}
		s.SetSessionID(payload.SessionId)
		service.SessionMapper.AddSessionId(payload.SessionId, s)
		fmt.Printf("add sessionid %v to stream %v", payload.SessionId, s.StreamId)
		reply := NewMsg([]byte("Reply add sessionid " + payload.SessionId + " to stream" + s.StreamId))
		_, err = s.RW.Write(reply)
		if err != nil {
			println("write err", err)
			return
		}
		err = s.RW.Flush()
		if err != nil {
			println("flush err", err)
			return
		}

	}
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
