package him

import (
	"context"
	"net"
	"time"
)

const (
	DefaultReadWait  = time.Minute * 3
	DefaultWriteWait = time.Second * 10
	DefaultLoginWait = time.Second * 10
	DefaultHeartbeat = time.Second * 55
)

// Service 定义基础服务的抽象接口
type Service interface {
	ServiceID() string
	ServiceName() string
	GetMeta() map[string]string
}

// ServiceRegistration Service define a Service
type ServiceRegistration interface {
	Service
	// PublicAddress ip or domain
	PublicAddress() string
	PublicPort() int
	DialURL() string
	GetProtocol() string
	GetNamespace() string
	GetTags() []string
	// String SetTags(tags []string)
	// SetMeta(meta map[string]string)
	String() string
}

// Server 服务器用于承载Service
type Server interface {
	ServiceRegistration
	SetAcceptor(Acceptor)
	SetMessageListener(MessageListener)
	SetStateListener(StateListener)
	SetReadWait(time.Duration)
	SetChannelMap(ChannelMap)

	Start() error
	Push(string, []byte) error
	Shutdown(ctx context.Context) error
}

// OpCode 定义统一的OpCode
type OpCode byte

const (
	OpContinuation OpCode = 0x0
	OpText         OpCode = 0x1
	OpBinary       OpCode = 0x2
	OpClose        OpCode = 0x8
	OpPing         OpCode = 0x9
	OpPong         OpCode = 0xa
)

// Frame 通过抽象一个Frame接口解决底层封包与拆包
type Frame interface {
	SetOpCode(OpCode)
	GetOpCode() OpCode
	SetPayload([]byte)
	GetPayload() []byte
}

// Conn 对net.Conn进行二次封装，把读与写的操作封装到连接中
type Conn interface {
	net.Conn
	ReadFrame() (Frame, error)
	WriteFrame(OpCode, []byte) error
	Flush() error
}

// Agent 表示发送方
type Agent interface {
	// ID 返回连接的channelID
	ID() string
	// Push 用于上层业务返回消息
	Push([]byte) error
}

// MessageListener 消息监听器
type MessageListener interface {
	// Receive 参数Agent表示发送方
	Receive(Agent, []byte)
}

// Acceptor 连接接收器
type Acceptor interface {
	// Accept 返回一个握手完成的Channel对象或者一个error。
	// 业务层需要处理不同协议和网络环境的下连接握手协议
	Accept(Conn, time.Duration) (string, error)
}

// StateListener 状态监听器
type StateListener interface {
	// Disconnect 连接断开回调
	Disconnect(string) error
}

type Handler interface {
	MessageListener
	Acceptor
	StateListener
}
