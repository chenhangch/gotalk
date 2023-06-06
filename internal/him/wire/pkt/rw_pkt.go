package pkt

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"reflect"

	"github.com/chang144/gotalk/internal/him/wire"
)

type Packet interface {
	Decode(r io.Reader) error
	Encode(r io.Writer) error
}

// ReadMagic 解包，根据魔数选择不同的协议
func ReadMagic(r io.Reader) (interface{}, error) {
	magic := wire.Magic{}
	_, err := io.ReadFull(r, magic[:])
	if err != nil {
		return nil, err
	}
	switch magic {
	case wire.MagicBasicPkt:
		p := new(HeartbeatPkt)
		if err := p.Decode(r); err != nil {
			return nil, err
		}
		return p, nil
	case wire.MagicLogicPkt:
		p := new(LogicPkt)
		if err := p.Decode(r); err != nil {
			return nil, err
		}
		return p, nil
	default:
		return nil, errors.New("magic backend is incorrect")
	}
}

// Marshal 封包，把Magic封装到消息的头部
func Marshal(p Packet) []byte {
	buf := new(bytes.Buffer)
	kind := reflect.TypeOf(p).Elem()
	if kind.AssignableTo(reflect.TypeOf(LogicPkt{})) {
		_, _ = buf.Write(wire.MagicLogicPkt[:])
	} else if kind.AssignableTo(reflect.TypeOf(HeartbeatPkt{})) {
		_, _ = buf.Write(wire.MagicBasicPkt[:])
	}
	err := p.Encode(buf)
	if err != nil {
		return nil
	}
	return buf.Bytes()
}

func MustReadLogicPkt(r io.Reader) (*LogicPkt, error) {
	val, err := ReadMagic(r)
	if err != nil {
		return nil, err
	}
	if lp, ok := val.(*LogicPkt); ok {
		return lp, nil
	}
	return nil, fmt.Errorf("packet is not a logic packet")
}

func MustReadBasicPkt(r io.Reader) (*HeartbeatPkt, error) {
	val, err := ReadMagic(r)
	if err != nil {
		return nil, err
	}
	if bp, ok := val.(*HeartbeatPkt); ok {
		return bp, nil
	}
	return nil, fmt.Errorf("packet is not a basic packet")
}
