package websocket

import (
	"net"

	"github.com/chang144/gotalk/internal/him"
	"github.com/gobwas/ws"
)

// Frame Frame 帧处理
// 对ws.Frame包装, 实现了him.Frame接口
type Frame struct {
	raw ws.Frame
}

func (f *Frame) SetOpCode(code him.OpCode) {
	f.raw.Header.OpCode = ws.OpCode(code)
}

func (f *Frame) GetOpCode() him.OpCode {
	return him.OpCode(f.raw.Header.OpCode)
}

func (f *Frame) SetPayload(payload []byte) {
	// note: 没有使用Mask编码
	// client 发送消息时不能直接使用websocket.Conn
	f.raw.Payload = payload
}

func (f *Frame) GetPayload() []byte {
	// 判断payload是否需要解码
	if f.raw.Header.Masked {
		ws.Cipher(f.raw.Payload, f.raw.Header.Mask, 0)
	}
	f.raw.Header.Masked = false
	return f.raw.Payload
}

// WsConn 对net.Conn进行封装
// 解决websocket/tcp两种协议在 读/写 逻辑上的差异
// 依赖 Frame接口的实现
type WsConn struct {
	net.Conn
}

func NewConn(conn net.Conn) *WsConn {
	return &WsConn{
		conn,
	}
}

func (c *WsConn) ReadFrame() (him.Frame, error) {
	frame, err := ws.ReadFrame(c.Conn)
	if err != nil {
		return nil, err
	}
	return &Frame{raw: frame}, nil
}

func (c *WsConn) WriteFrame(code him.OpCode, payload []byte) error {
	// fin : true --> 发送的数据报不能超过websocket单个帧最大值
	f := ws.NewFrame(ws.OpCode(code), true, payload)
	return ws.WriteFrame(c.Conn, f)
}

func (c *WsConn) Flush() error {
	return nil
}
