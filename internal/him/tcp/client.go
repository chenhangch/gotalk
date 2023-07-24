package tcp

import (
	"errors"
	"fmt"
	"net/url"
	"sync"
	"sync/atomic"
	"time"

	"github.com/chang144/gotalk/internal/him"
)

type ClientOptions struct {
	Heartbeat time.Duration //登录超时
	ReadWait  time.Duration //读超时
	WriteWait time.Duration //写超时
}

// Client is a tcp implement of kim.Client
type Client struct {
	sync.Mutex
	him.Dialer
	once    sync.Once
	id      string
	name    string
	conn    him.Conn
	state   int32
	options ClientOptions

	Meta map[string]string
}

func NewClient(id, name string, opts ClientOptions) him.Client {
	return NewClientWithProps(id, name, make(map[string]string), opts)
}

func NewClientWithProps(id, name string, meta map[string]string, opts ClientOptions) him.Client {
	if opts.WriteWait == 0 {
		opts.WriteWait = him.DefaultWriteWait
	}
	if opts.ReadWait == 0 {
		opts.ReadWait = him.DefaultReadWait
	}

	cli := &Client{
		id:      id,
		name:    name,
		options: opts,
		Meta:    meta,
	}
	return cli
}

func (c *Client) ServiceID() string {
	return c.id
}

func (c *Client) ServiceName() string {
	return c.name
}

func (c *Client) GetMeta() map[string]string {
	return c.Meta
}

func (c *Client) SetDialer(dialer him.Dialer) {
	c.Dialer = dialer
}

// Send data to connection
func (c *Client) Send(payload []byte) error {
	if atomic.LoadInt32(&c.state) == 0 {
		return fmt.Errorf("connection is nil")
	}
	c.Lock()
	defer c.Unlock()
	_ = c.conn.SetWriteDeadline(time.Now().Add(c.options.WriteWait))
	return c.conn.WriteFrame(him.OpBinary, payload)
}

func (c *Client) Read() (him.Frame, error) {
	if c.conn == nil {
		return nil, errors.New("connection is nil")
	}
	if c.options.Heartbeat > 0 {
		_ = c.conn.SetReadDeadline(time.Now().Add(c.options.ReadWait))
	}
	frame, err := c.conn.ReadFrame()
	if err != nil {
		return nil, err
	}
	if frame.GetOpCode() == him.OpClose {
		return nil, errors.New("remote side close the channel")
	}
	return frame, nil
}

// Close 关闭
func (c *Client) Close() {
	c.once.Do(func() {
		if c.conn == nil {
			_ = WriteFrame(c.conn, him.OpClose, nil)

			c.conn.Close()
			atomic.CompareAndSwapInt32(&c.state, 1, 0)
		}
	})
}

// Connect to logicServer
func (c *Client) Connect(addr string) error {
	_, err := url.Parse(addr)
	if err != nil {
		return err
	}
	// CAS原子操作，对比并设置值，并发安全
	if !atomic.CompareAndSwapInt32(&c.state, 0, 1) {
		return fmt.Errorf("client has connected")
	}

	rawconn, err := c.Dialer.DialAndHandshake(him.DialerContext{
		Id:      c.id,
		Name:    c.name,
		Address: addr,
		Timeout: him.DefaultLoginWait,
	})
	if err != nil {
		atomic.CompareAndSwapInt32(&c.state, 1, 0)
		return err
	}
	if rawconn == nil {
		return fmt.Errorf("conn is nil")
	}
	c.conn = NewConn(rawconn)

	if c.options.Heartbeat > 0 {
		go func() {
			err := c.heartbeatLoop()
			if err != nil {
				return
			}
		}()
	}
	return nil
}

func (c *Client) heartbeatLoop() error {
	ticker := time.NewTicker(c.options.Heartbeat)
	for range ticker.C {
		if err := c.ping(); err != nil {
			return err
		}
	}
	return nil
}

func (c *Client) ping() error {
	err := c.conn.SetWriteDeadline(time.Now().Add(c.options.WriteWait))
	if err != nil {
		return err
	}
	return c.conn.WriteFrame(him.OpPing, nil)
}
