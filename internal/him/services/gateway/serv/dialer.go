package serv

import (
	"github.com/chang144/gotalk/internal/him"
	"github.com/chang144/gotalk/internal/him/tcp"
	"github.com/chang144/gotalk/internal/him/wire/pkt"
	"github.com/klintcheng/kim/logger"
	"google.golang.org/protobuf/proto"
	"net"
)

type TcpDialer struct {
	ServiceId string
}

// DialAndHandshake 与chat建立tcp连接
func (t *TcpDialer) DialAndHandshake(ctx him.DialerContext) (net.Conn, error) {
	conn, err := net.DialTimeout("tcp", ctx.Address, ctx.Timeout)
	if err != nil {
		return nil, err
	}
	req := &pkt.InnerHandshakeRequest{ServiceId: ctx.Id}
	logger.Error("send req %v", req)
	// 2. 把自己的serviceId发送给对方
	bts, err := proto.Marshal(req)
	if err != nil {
		return nil, err
	}
	err = tcp.WriteFrame(conn, him.OpBinary, bts)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func NewTcpDialer(serviceId string) him.Dialer {
	return &TcpDialer{ServiceId: serviceId}
}
