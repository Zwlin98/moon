package cluster

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

type Response struct {
	Ok      bool
	Session uint32
	Msg     []byte

	Padding uint8 // for received response
}

type PackedResponse struct {
	Data  []byte
	Multi [][]byte
}

func PackResponse(r Response) (PackedResponse, error) {
	if !r.Ok {
		return packSingleResponse(r)
	}
	msgSize := uint32(len(r.Msg))
	if msgSize > MULTI_PART {
		return packMultiResponse(r)
	}
	return packSingleResponse(r)
}

func UnpackResponse(data []byte) (Response, error) {
	if len(data) < 5 {
		return Response{}, fmt.Errorf("response data is too short")
	}
	var r = Response{}
	r.Session = binary.LittleEndian.Uint32(data)
	switch data[4] {
	case RESPONSE_OK:
		r.Ok = true
		r.Msg = data[5:]
		r.Padding = RESPONSE_END
	case RESPONSE_ERROR:
		r.Ok = false
		r.Msg = data[5:]
		r.Padding = RESPONSE_MULTI_END
	case RESPONSE_MULTI_BEGIN:
		r.Ok = true
		r.Padding = RESPONSE_MULTI_BEGIN
	case RESPONSE_MULTI_PART:
		r.Msg = data[5:]
		r.Padding = RESPONSE_MULTI_PART
	case RESPONSE_MULTI_END:
		r.Msg = data[5:]
		r.Padding = RESPONSE_MULTI_END
	default:
		return r, fmt.Errorf("response type is not supported")
	}
	return r, nil
}

func packSingleResponse(r Response) (PackedResponse, error) {
	var pr = PackedResponse{}
	buf := bytes.NewBuffer(nil)

	// 4 bytes for session
	binary.Write(buf, binary.LittleEndian, r.Session)
	// 1 byte for request type
	if r.Ok {
		binary.Write(buf, binary.LittleEndian, RESPONSE_OK)
	} else {
		binary.Write(buf, binary.LittleEndian, RESPONSE_ERROR)
		// truncate err message if it's too long
		if len(r.Msg) > MULTI_PART {
			r.Msg = r.Msg[:MULTI_PART]
		}
	}

	buf.Write(r.Msg)

	pr.Data = buf.Bytes()
	return pr, nil
}

func packMultiResponse(r Response) (PackedResponse, error) {
	var pr = PackedResponse{
		Data:  make([]byte, 9),
		Multi: make([][]byte, 0),
	}
	msgSize := uint32(len(r.Msg))

	// 4 bytes for session
	binary.LittleEndian.PutUint32(pr.Data, r.Session)
	// 1 byte for request type
	pr.Data[4] = RESPONSE_MULTI_BEGIN
	// 4 bytes for msg size
	binary.LittleEndian.PutUint32(pr.Data[5:], msgSize)

	part := int((msgSize-1)/MULTI_PART + 1)
	for i := 0; i < part; i++ {
		bufPart := bytes.NewBuffer(nil)
		var s uint32
		var respType uint8
		if msgSize > MULTI_PART {
			s = MULTI_PART
			respType = RESPONSE_MULTI_PART
		} else {
			s = msgSize
			respType = RESPONSE_MULTI_END
		}
		// 4 bytes for session
		binary.Write(bufPart, binary.LittleEndian, r.Session)
		// 1 byte for request type
		bufPart.WriteByte(respType)

		partStart := i * int(MULTI_PART)
		partEnd := partStart + int(s)

		// copy msg
		bufPart.Write(r.Msg[partStart:partEnd])

		pr.Multi = append(pr.Multi, bufPart.Bytes())

		msgSize -= s
	}

	return pr, nil
}
