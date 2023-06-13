package serv

import (
	"bytes"
	"fmt"
	"github.com/chang144/gotalk/internal/him/container"
	"github.com/chang144/gotalk/internal/him/wire"
	"github.com/chang144/gotalk/internal/him/wire/pkt"
	"github.com/chang144/gotalk/internal/him/wire/token"
	"regexp"

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
	AppSecret string
}

// Receive 接收SDK发送来的消息
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
	log.Info("disconnect %s", id)
	logout := pkt.New(wire.CommandLoginSignOut, pkt.WithChannel(id))
	err := container.Forward(wire.SNLogin, logout)
	if err != nil {
		logger.WithFields(logger.Fields{
			"module": "handler",
			"id":     id,
		}).Error(err)
	}
	return nil
}

func (h *Handler) Accept(conn him.Conn, timeout time.Duration) (string, error) {
	// TODO 日志记录
	_ = conn.SetReadDeadline(time.Now().Add(timeout))
	// 读取登录包
	frame, err := conn.ReadFrame()
	if err != nil {
		return "", err
	}

	buffer := bytes.NewBuffer(frame.GetPayload())
	req, err := pkt.MustReadLogicPkt(buffer)
	if err != nil {
		return "", err
	}
	// 判断是不是登录包
	if req.Command != wire.CommandLoginSignIn {
		resp := pkt.NewLogicPkt(&req.Header)
		resp.Status = pkt.Status_InvalidCommand
		_ = conn.WriteFrame(him.OpBinary, pkt.Marshal(resp))
		return "", fmt.Errorf("must be a InvalidCommand command")
	}

	// 反序列化body
	var login pkt.LoginReq
	err = req.ReadBody(&login)
	if err != nil {
		return "", err
	}

	// 使用默认的DefaultSecret解析token
	tk, err := token.Parse(token.DefaultSecret, login.Token)
	if err != nil {
		// token 无效
		resp := pkt.NewLogicPkt(&req.Header)
		resp.Status = pkt.Status_Unauthorized
		_ = conn.WriteFrame(him.OpBinary, pkt.Marshal(resp))
		return "", err
	}
	// 生成一个全局唯一的ChannelID
	id := fmt.Sprintf("%s_%s_%d", h.ServiceId, tk.Account, wire.Seq.Next())

	req.ChannelId = id
	req.WriteBody(&pkt.Session{
		ChannelId: id,
		GateId:    h.ServiceId,
		Account:   tk.Account,
		RemoteIP:  getIP(conn.RemoteAddr().String()),
		App:       tk.App,
	})
	// 7. 把login.转发给Login服务
	err = container.Forward(wire.SNLogin, req)
	if err != nil {
		return "", err
	}
	return id, nil
}

var _ him.Handler = (*Handler)(nil)

var ipExp = regexp.MustCompile(string("\\:[0-9]+$"))

func getIP(remoteAddr string) string {
	if remoteAddr == "" {
		return ""
	}
	return ipExp.ReplaceAllString(remoteAddr, "")
}
