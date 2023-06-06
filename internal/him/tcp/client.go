package tcp

import (
	"sync"
	"time"

	"github.com/chang144/gotalk/internal/him"
	"github.com/klintcheng/kim"
)

type ClientOptions struct {
	Heartbeat time.Duration //登录超时
	ReadWait  time.Duration //读超时
	WriteWait time.Duration //写超时
}

// Client is a tcp implement of kim.Client
type Client struct {
	sync.Mutex
	kim.Dialer
	once    sync.Once
	id      string
	name    string
	conn    kim.Conn
	state   int32
	options ClientOptions

	Meta map[string]string
}

func (c *Client) ServiceID() string {
	//TODO implement me
	panic("implement me")
}

func (c *Client) ServiceName() string {
	//TODO implement me
	panic("implement me")
}

func (c *Client) GetMeta() map[string]string {
	//TODO implement me
	panic("implement me")
}

func (c *Client) ID() string {
	//TODO implement me
	panic("implement me")
}

func (c *Client) Name() string {
	//TODO implement me
	panic("implement me")
}

func (c *Client) SetDialer(dialer him.Dialer) {
	//TODO implement me
	panic("implement me")
}

func (c *Client) Send(bytes []byte) error {
	//TODO implement me
	panic("implement me")
}

func (c *Client) Read() (him.Frame, error) {
	//TODO implement me
	panic("implement me")
}

func (c *Client) Close() {
	//TODO implement me
	panic("implement me")
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

// Connect to chatServer
func (c *Client) Connect(addr string) error {

	return nil
}
