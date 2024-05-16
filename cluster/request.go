package cluster

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

type Request struct {
	Address any // uint32 or string
	Session uint32
	IsPush  bool
	Msg     []byte

	Completed bool // for received request
}

type PackedRequest struct {
	Data  []byte
	Multi [][]byte
}

func PackRequest(r Request) (PackedRequest, error) {
	var pr = PackedRequest{}
	if r.Address == nil {
		return pr, fmt.Errorf("address is nil")
	}
	if r.Session == 0 {
		return pr, fmt.Errorf("session is zero")
	}
	msgSize := uint32(len(r.Msg))
	if msgSize == 0 {
		return pr, fmt.Errorf("msg is empty")
	}
	if msgSize < MULTI_PART {
		return packSingleRequest(r)
	}
	return packMultiRequest(r)
}

func UnpackRequest(data []byte) (Request, error) {
	var r = Request{}
	switch data[0] {
	case REQUEST_SINGLE_NUMBER:
		fallthrough
	case REQUEST_SINGLE_STRING:
		return unpackSingleRequest(data)
	case REQUEST_MULTI_NUMBER:
		fallthrough
	case REQUEST_MULTI_NUMBER_PUSH:
		fallthrough
	case REQUEST_MULTI_STRING:
		fallthrough
	case REQUEST_MULTI_STRING_PUSH:
		fallthrough
	case REQUEST_MULTI_PART:
		fallthrough
	case REQUEST_MULTI_PART_END:
		return unpackMultiRequest(data)
	default:
		return r, fmt.Errorf("request type is not supported")
	}
}

func unpackSingleRequest(data []byte) (Request, error) {
	var r = Request{}
	switch data[0] {
	case REQUEST_SINGLE_NUMBER:
		r.Address = binary.LittleEndian.Uint32(data[1:])
		r.Session = binary.LittleEndian.Uint32(data[5:])
		r.Msg = data[9:]
	case REQUEST_SINGLE_STRING:
		nameLen := uint8(data[1])
		r.Address = string(data[2 : 2+nameLen])
		r.Session = binary.LittleEndian.Uint32(data[2+nameLen:])
		r.Msg = data[6+nameLen:]
	}
	if r.Session == 0 {
		r.IsPush = true
	} else {
		r.IsPush = false
	}
	r.Completed = true
	return r, nil
}

func unpackMultiRequest(data []byte) (Request, error) {
	var r = Request{}
	switch data[0] {
	case REQUEST_MULTI_NUMBER, REQUEST_MULTI_NUMBER_PUSH:
		if data[0] == REQUEST_MULTI_NUMBER_PUSH {
			r.IsPush = true
		} else {
			r.IsPush = false
		}
		r.Address = binary.LittleEndian.Uint32(data[1:])
		r.Session = binary.LittleEndian.Uint32(data[5:])
	case REQUEST_MULTI_STRING, REQUEST_MULTI_STRING_PUSH:
		if data[0] == REQUEST_MULTI_STRING_PUSH {
			r.IsPush = true
		} else {
			r.IsPush = false
		}
		nameLen := uint8(data[1])
		r.Address = string(data[2 : 2+nameLen])
		r.Session = binary.LittleEndian.Uint32(data[2+nameLen:])
	case REQUEST_MULTI_PART, REQUEST_MULTI_PART_END:
		if data[0] == REQUEST_MULTI_PART_END {
			r.Completed = true
		}
		r.Session = binary.LittleEndian.Uint32(data[1:])
		r.Msg = data[5:]
	}
	return r, nil
}

func packSingleRequest(r Request) (PackedRequest, error) {
	var pr = PackedRequest{}
	var buf = bytes.NewBuffer(nil)
	switch r.Address.(type) {
	case uint32:
		address := r.Address.(uint32)
		// 1 byte for request type
		buf.WriteByte(REQUEST_SINGLE_NUMBER)
		// 4 bytes for address
		binary.Write(buf, binary.LittleEndian, address)
		// 4 bytes for session
		if !r.IsPush {
			binary.Write(buf, binary.LittleEndian, r.Session)
		} else {
			binary.Write(buf, binary.LittleEndian, uint32(0))
		}
		// copy msg
		buf.Write(r.Msg)
	case string:
		name := r.Address.(string)
		nameLen := uint8(len(name))
		if nameLen < 1 || nameLen > 255 {
			return PackedRequest{}, fmt.Errorf("name length error")
		}
		// 1 byte for request type
		buf.WriteByte(REQUEST_SINGLE_STRING)
		// 1 byte for name length
		buf.WriteByte(uint8(nameLen))
		// `nameLen` bytes for name
		buf.Write([]byte(name))
		// 4 bytes for session
		if !r.IsPush {
			binary.Write(buf, binary.LittleEndian, r.Session)
		} else {
			binary.Write(buf, binary.LittleEndian, uint32(0))
		}
		// copy msg
		buf.Write(r.Msg)
	default:
		return PackedRequest{}, fmt.Errorf("address type is not supported")
	}
	pr.Data = buf.Bytes()
	return pr, nil
}

func packMultiRequest(r Request) (PackedRequest, error) {
	var pr = PackedRequest{
		Multi: make([][]byte, 0),
	}
	msgSize := uint32(len(r.Msg))

	bufData := bytes.NewBuffer(nil)
	switch r.Address.(type) {
	case uint32:
		address := r.Address.(uint32)
		// 1 byte for request type
		if r.IsPush {
			bufData.WriteByte(REQUEST_MULTI_NUMBER_PUSH)
		} else {
			bufData.WriteByte(REQUEST_MULTI_NUMBER)
		}
		// 4 bytes for address
		binary.Write(bufData, binary.LittleEndian, address)
		// 4 bytes for session
		binary.Write(bufData, binary.LittleEndian, r.Session)
		// 4 bytes for msg size
		binary.Write(bufData, binary.LittleEndian, msgSize)
	case string:
		name := r.Address.(string)
		nameLen := uint16(len(name))
		if nameLen < 1 || nameLen > 255 {
			return PackedRequest{}, fmt.Errorf("name length error")
		}
		// 1 byte for request type
		if r.IsPush {
			bufData.WriteByte(REQUEST_MULTI_STRING_PUSH)
		} else {
			bufData.WriteByte(REQUEST_MULTI_STRING)
		}
		// 1 byte for name length
		bufData.WriteByte(uint8(nameLen))
		// `nameLen` bytes for name
		bufData.Write([]byte(name))
		// 4 bytes for session
		binary.Write(bufData, binary.LittleEndian, r.Session)
		// 4 bytes for msg size
		binary.Write(bufData, binary.LittleEndian, msgSize)
	default:
		return PackedRequest{}, fmt.Errorf("address type is not supported")
	}
	pr.Data = bufData.Bytes()

	part := int((msgSize-1)/MULTI_PART + 1)
	for i := 0; i < part; i++ {
		partBuf := bytes.NewBuffer(nil)
		var s uint32
		var reqType uint8
		if msgSize > MULTI_PART {
			s = MULTI_PART
			reqType = REQUEST_MULTI_PART
		} else {
			s = msgSize
			reqType = REQUEST_MULTI_PART_END
		}
		partBuf.WriteByte(reqType)
		// 4 bytes for session
		binary.Write(partBuf, binary.LittleEndian, r.Session)

		partStart := i * int(MULTI_PART)
		partEnd := partStart + int(s)

		// copy msg
		partBuf.Write(r.Msg[partStart:partEnd])

		pr.Multi = append(pr.Multi, partBuf.Bytes())
		msgSize -= s
	}

	return pr, nil
}
