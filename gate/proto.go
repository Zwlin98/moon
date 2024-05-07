package gate

import (
	"encoding/binary"
	"fmt"
	"io"
)

type GateProto interface {
	ReadMsg() ([]byte, error)
	WriteMsg([]byte) error
}

type skynetGateProto struct {
	reader io.Reader
	writer io.Writer
}

func NewGateProto(rd io.Reader, wd io.Writer) GateProto {
	return &skynetGateProto{
		reader: rd,
		writer: wd,
	}
}

func (gp *skynetGateProto) ReadMsg() ([]byte, error) {
	var sz uint16
	err := binary.Read(gp.reader, binary.BigEndian, &sz)
	if err != nil {
		return nil, err
	}
	buf := make([]byte, sz)
	_, err = io.ReadFull(gp.reader, buf)
	if err != nil {
		return nil, err
	}
	return buf, nil
}

func (gp *skynetGateProto) WriteMsg(msg []byte) error {
	if len(msg) > 0x10000 {
		return fmt.Errorf("message too long")
	}
	sz := uint16(len(msg))
	err := binary.Write(gp.writer, binary.BigEndian, sz)
	if err != nil {
		return err
	}
	_, err = gp.writer.Write(msg)
	return err
}
