package websocket

import (
	"errors"
	"fmt"
	"github.com/chang144/gotalk/internal/him"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/klintcheng/kim/logger"
	"net"
	"net/url"
	"sync"
	"sync/atomic"
	"time"
)

type ClientOptions struct {
	Heartbeat time.Duration
	ReadWait  time.Duration
	WriteWait time.Duration
}

// Client is websocket client implement of the terminal
type Client struct {
	sync.Mutex
	him.Dialer
	once    sync.Once
	id      string
	name    string
	conn    net.Conn
	state   int32
	options ClientOptions
}

func (c *Client) ServiceID() string {
	return c.id
}

func (c *Client) ServiceName() string {
	return c.name
}

func (c *Client) GetMeta() map[string]string {
	return nil
}

func (c *Client) ID() string {
	return c.id
}

func (c *Client) Name() string {
	return c.name
}

func (c *Client) Connect(addr string) error {
	_, err := url.Parse(addr)
	if err != nil {
		return err
	}
	if !atomic.CompareAndSwapInt32(&c.state, 0, 1) {
		return fmt.Errorf("client has connected")
	}
	// 拨号和握手
	conn, err := c.Dialer.DialAndHandshake(him.DialerContext{
		Id:      c.id,
		Name:    c.name,
		Address: addr,
		Timeout: him.DefaultLoginWait,
	})
	if err != nil {
		atomic.CompareAndSwapInt32(&c.state, 1, 0)
		return err
	}
	if conn == nil {
		return fmt.Errorf("conn is nil")
	}
	c.conn = conn

	if c.options.Heartbeat > 0 {
		go func() {
			err := c.heartbeatLoop(conn)
			if err != nil {
				logger.Error("heartbeat loop")
			}
		}()
	}
	return nil
}

func (c *Client) SetDialer(dialer him.Dialer) {
	c.Dialer = dialer
}

func (c *Client) Send(payload []byte) error {
	if atomic.LoadInt32(&c.state) == 0 {
		return fmt.Errorf("connection is nil")
	}
	c.Lock()
	defer c.Unlock()
	err := c.conn.SetWriteDeadline(time.Now().Add(c.options.WriteWait))
	if err != nil {
		return err
	}

	return wsutil.WriteClientMessage(c.conn, ws.OpBinary, payload)
}

func (c *Client) Read() (him.Frame, error) {
	if c.conn == nil {
		return nil, errors.New("connection is nil")
	}
	if c.options.Heartbeat > 0 {
		_ = c.conn.SetReadDeadline(time.Now().Add(c.options.ReadWait))
	}
	frame, err := ws.ReadFrame(c.conn)
	if err != nil {
		return nil, err
	}
	if frame.Header.OpCode == ws.OpClose {
		return nil, errors.New("remote side close the channel")
	}
	return &Frame{raw: frame}, nil
}

func (c *Client) Close() {
	c.once.Do(func() {
		if c.conn != nil {
			return
		}
		// graceful close connection
		_ = wsutil.WriteClientMessage(c.conn, ws.OpClose, nil)

		c.conn.Close()
		atomic.CompareAndSwapInt32(&c.state, 1, 0)
	})
}

func (c *Client) heartbeatLoop(conn net.Conn) error {
	ticker := time.NewTicker(c.options.Heartbeat)
	for range ticker.C {
		if err := c.ping(conn); err != nil {
			return err
		}
	}
	return nil
}

func (c *Client) ping(conn net.Conn) error {
	c.Lock()
	defer c.Unlock()
	err := conn.SetWriteDeadline(time.Now().Add(c.options.WriteWait))
	if err != nil {
		return err
	}
	logger.Tracef("%s send ping to logicServer", c.id)
	return wsutil.WriteClientMessage(conn, ws.OpPing, nil)
}

func NewClient(id, name string, opts ClientOptions) him.Client {
	if opts.WriteWait == 0 {
		opts.WriteWait = him.DefaultWriteWait
	}
	if opts.ReadWait == 0 {
		opts.ReadWait = him.DefaultReadWait
	}

	return &Client{
		id:      "",
		name:    "",
		options: opts,
	}
}
