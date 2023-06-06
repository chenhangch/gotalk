package pkt

import (
	"github.com/klintcheng/kim/wire/endian"
	"io"
)

// basic packet backend
const (
	CodePing = uint16(1)
	CodePong = uint16(2)
)

// HeartbeatPkt basic pkt backend
type HeartbeatPkt struct {
	Code   uint16
	Length uint16
	Body   []byte
}

// Decode 解码
func (p *HeartbeatPkt) Decode(r io.Reader) error {
	var err error
	if p.Length, err = endian.ReadUint16(r); err != nil {
		return err
	}
	if p.Length > 0 {
		if p.Body, err = endian.ReadFixedBytes(int(p.Length), r); err != nil {
			return err
		}
	}
	return nil
}

// Encode 加密
func (p *HeartbeatPkt) Encode(w io.Writer) error {
	if err := endian.WriteUint16(w, p.Code); err != nil {
		return err
	}
	if err := endian.WriteUint16(w, p.Length); err != nil {
		return err
	}
	if p.Length > 0 {
		if _, err := w.Write(p.Body); err != nil {
			return err
		}
	}
	return nil
}
