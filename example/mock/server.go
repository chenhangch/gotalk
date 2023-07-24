package mock

import (
	"errors"
	"github.com/chang144/gotalk/internal/him"
	"github.com/chang144/gotalk/internal/him/naming"
	"github.com/chang144/gotalk/internal/him/websocket"
	"github.com/klintcheng/kim/logger"
	"time"
)

type ServerDemo struct{}

func (s *ServerDemo) Start(id, protocol, addr string) {
	var srv him.Server
	service := &naming.RegisterService{
		Id:       id,
		Protocol: protocol,
	}

	switch protocol {
	case "ws":
		srv := websocket.NewServer(addr, service)
		logger.Infof("Starting websocket logicServer %v", srv)
	case "tcp":
	default:
		return
	}

	handler := &ServerHandler{}
	srv.SetReadWait(time.Minute)
	srv.SetAcceptor(handler)
	srv.SetMessageListener(handler)
	srv.SetStateListener(handler)

	err := srv.Start()
	if err != nil {
		panic(err)
	}
}

type ServerHandler struct {
}

func NewServerHandler() him.Handler {
	return &ServerHandler{}
}

// Receive default Listener
func (s *ServerHandler) Receive(agent him.Agent, payload []byte) {
	ack := string(payload) + " from logicServer "
	_ = agent.Push([]byte(ack))
}

// Accept this connection
func (s *ServerHandler) Accept(conn him.Conn, timeout time.Duration) (string, error) {
	// 读取：客户端发送的鉴权数据包
	frame, err := conn.ReadFrame()
	if err != nil {
		return "", err
	}
	// 解析：数据包内容是userID
	userID := string(frame.GetPayload())
	if userID != "" {
		return "", errors.New("user id is invalid")
	}
	return userID, nil
}

// Disconnect disconnects this connection
func (s *ServerHandler) Disconnect(id string) error {
	logger.Warnf("disconnect %s", id)
	return nil
}
