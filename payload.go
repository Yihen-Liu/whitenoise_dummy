package whitenoise

import (
	"encoding/binary"
	"github.com/golang/protobuf/proto"
	"whitenoise/pb"
)

const TYPELEN = 4
const LENLEN = 4
const MaxLen = 1000

func EncodePayload(data []byte) []byte {
	l := uint32(len(data))
	b := Int2Bytes(l)
	return append(b[:], data...)
}

func Int2Bytes(l uint32) [LENLEN]byte {
	var res [LENLEN]byte
	binary.BigEndian.PutUint32(res[:], l)
	return res
}

func Bytes2Int(b []byte) uint32 {
	return binary.BigEndian.Uint32(b)
}

func NewSetSessionIDCommand(sessionID string, streamID string) []byte {
	payload := pb.Payload{
		SessionId: sessionID,
		StreamId:  streamID,
		Data:      []byte{},
	}
	comd, _ := proto.Marshal(&payload)
	return EncodePayload(comd)
}

func NewMsg(data []byte) []byte {
	payload := pb.Payload{
		SessionId: "",
		StreamId:  "",
		Data:      data,
	}
	comd, _ := proto.Marshal(&payload)
	return EncodePayload(comd)
}
