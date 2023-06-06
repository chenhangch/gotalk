package serv

import (
	"bytes"
	"github.com/chang144/gotalk/internal/him/container"
	"github.com/chang144/gotalk/internal/him/wire/pkt"

	"time"

	"github.com/chang144/gotalk/internal/him"
	"github.com/klintcheng/kim/logger"
)

var log = logger.WithFields(logger.Fields{
	"service": "gateway",
	"pkg":     "serv",
})

type Handler struct {
	ServiceId string
}

func (h *Handler) Receive(agent him.Agent, payload []byte) {
	buf := bytes.NewBuffer(payload)
	packet, err := pkt.ReadMagic(buf)
	if err != nil {
		return
	}

	// 如果是basicPkt 就处理心跳包
	if hPkt, ok := packet.(*pkt.HeartbeatPkt); ok {
		if hPkt.Code == pkt.CodePing {
			_ = agent.Push(pkt.Marshal(&pkt.HeartbeatPkt{Code: pkt.CodePong}))
		}
		return
	}
	// 如果是LoginPkt，就转化给逻辑处理服务器
	if logicPkt, ok := packet.(*pkt.LogicPkt); ok {
		logicPkt.ChannelId = agent.ID()

		err := container.Forward(logicPkt.ServiceName(), logicPkt)
		if err != nil {
			logger.WithFields(logger.Fields{
				"module": "handler",
				"id":     agent.ID(),
				"cmd":    logicPkt.Command,
				"dest":   logicPkt.Dest,
			}).Error(err)
		}

	}

}

func (h *Handler) Disconnect(id string) error {
	//TODO implement me
	panic("implement me")
}

func (h *Handler) Accept(conn him.Conn, timeout time.Duration) (string, error) {
	//TODO implement me
	panic("implement me")
}

var _ him.Handler = (*Handler)(nil)
