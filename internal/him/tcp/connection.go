package tcp

import (
	"github.com/chang144/gotalk/internal/him"
	"github.com/klintcheng/kim/wire/endian"
	"io"
	"net"
)

// Frame tcp frame
type Frame struct {
	OpCode  him.OpCode
	Payload []byte
}

type TcpConn struct {
	net.Conn
}

func NewConn(conn net.Conn) *TcpConn {
	return &TcpConn{
		Conn: conn,
	}
}

func (c *TcpConn) ReadFrame() (him.Frame, error) {
	opcode, err := endian.ReadUint8(c.Conn)
	if err != nil {
		return nil, err
	}
	payload, err := endian.ReadBytes(c.Conn)
	if err != nil {
		return nil, err
	}
	return &Frame{
		OpCode:  him.OpCode(opcode),
		Payload: payload,
	}, nil
}

func (f *Frame) SetOpCode(code him.OpCode) {
	f.OpCode = code
}

func (f *Frame) GetOpCode() him.OpCode {
	return f.OpCode
}

func (f *Frame) SetPayload(payload []byte) {
	f.Payload = payload
}

func (f *Frame) GetPayload() []byte {
	return f.Payload
}

func (c *TcpConn) WriteFrame(code him.OpCode, payload []byte) error {
	return WriteFrame(c.Conn, code, payload)
}

func (c *TcpConn) Flush() error {
	return nil
}

func WriteFrame(w io.Writer, code him.OpCode, payload []byte) error {
	if err := endian.WriteUint8(w, uint8(code)); err != nil {
		return err
	}
	if err := endian.WriteBytes(w, payload); err != nil {
		return err
	}
	return nil
}
