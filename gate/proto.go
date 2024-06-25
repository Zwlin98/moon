package gate

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
)

type GateProto interface {
	Read() ([]byte, error)
	Write([]byte) error
	WriteBatch([][]byte) error
}

type skynetGateProto struct {
	reader *bufio.Reader
	writer *bufio.Writer
}

func NewGateProto(rd io.Reader, wd io.Writer) GateProto {
	return &skynetGateProto{
		reader: bufio.NewReader(rd),
		writer: bufio.NewWriter(wd),
	}
}

func (gp *skynetGateProto) Read() ([]byte, error) {
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

func (gp *skynetGateProto) writeMsg(msg []byte) error {
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

func (gp *skynetGateProto) Write(msg []byte) error {
	if err := gp.writeMsg(msg); err != nil {
		return err
	}
	return gp.writer.Flush()
}

func (gp *skynetGateProto) WriteBatch(msgs [][]byte) error {
	for _, msg := range msgs {
		if err := gp.writeMsg(msg); err != nil {
			return err
		}
	}
	return gp.writer.Flush()
}
