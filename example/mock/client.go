package mock

import (
	"context"
	"github.com/chang144/gotalk/internal/him"
	"github.com/chang144/gotalk/internal/him/websocket"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/klintcheng/kim/logger"
	"net"
	"time"
)

type ClientDemo struct {
}

func (c *ClientDemo) Start(userID, protocol, addr string) {
	var cli him.Client

	if protocol == "ws" {
		cli = websocket.NewClient(userID, "client", websocket.ClientOptions{
			Heartbeat: him.DefaultHeartbeat,
			ReadWait:  him.DefaultReadWait,
			WriteWait: him.DefaultWriteWait,
		})
		cli.SetDialer(&WebsocketDialer{})
	}

	err := cli.Connect(addr)
	if err != nil {
		logger.Error(err)
	}

	count := 10
	go func() {
		for i := 0; i < count; i++ {
			err := cli.Send([]byte("hello"))
			if err != nil {
				logger.Error(err)
				return
			}
			time.Sleep(time.Second)
		}
	}()

	recv := 0
	for {
		frame, err := cli.Read()
		if err != nil {
			break
		}

		if frame.GetOpCode() != him.OpBinary {
			continue
		}

		recv++
		logger.Warnf("%s receive message [%s]", cli.ID(), frame.GetPayload())
		if recv == count { // 接收完消息
			break
		}
	}

	cli.Close()
}

// WebsocketDialer WebsocketDialer
type WebsocketDialer struct {
}

// DialAndHandshake DialAndHandshake
func (d *WebsocketDialer) DialAndHandshake(ctx him.DialerContext) (net.Conn, error) {
	// 1 调用ws.Dial拨号
	conn, _, _, err := ws.Dial(context.TODO(), ctx.Address)
	if err != nil {
		return nil, err
	}
	// 2. 发送用户认证信息，示例就是userid
	err = wsutil.WriteClientBinary(conn, []byte(ctx.Id))
	if err != nil {
		return nil, err
	}
	// 3. return conn
	return conn, nil
}
