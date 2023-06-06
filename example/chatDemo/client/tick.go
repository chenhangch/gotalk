package client

import (
	"context"
	"errors"
	"fmt"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"net"
	"net/url"
	"time"
)

type StartOptions struct {
	user    string
	address string
}

// NewCmd NewCmd
func NewCmd(ctx context.Context) *cobra.Command {
	opts := &StartOptions{}

	cmd := &cobra.Command{
		Use:   "client",
		Short: "Start client",
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(ctx, opts)
		},
	}
	cmd.PersistentFlags().StringVarP(&opts.address, "address", "a", "ws://127.0.0.1:8000", "chatServer address")
	cmd.PersistentFlags().StringVarP(&opts.user, "user", "u", "", "user")
	return cmd
}

func run(ctx context.Context, opts *StartOptions) error {
	url := fmt.Sprintf("%s?user=%s", opts.address, opts.user)
	logrus.Infof("connect to %s", url)

	// 连接服务器 并返回handler对象
	h, err := connect(url)
	if err != nil {
		return err
	}

	go func() {
		for msg := range h.recv {
			logrus.Info("Received message", string(msg))
		}
	}()

	tk := time.NewTicker(time.Second * 6)
	for {
		select {
		case <-tk.C:
			err := h.SendText("hello")
			if err != nil {
				logrus.Error(err)
			}
		case <-h.close:
			logrus.Printf("connection closed")
			return nil
		}
	}

}

type handler struct {
	conn      net.Conn
	close     chan struct{}
	recv      chan []byte
	heartbeat time.Duration
}

func (h *handler) readLoop(conn net.Conn) error {
	logrus.Info("read loop started")

	err := h.conn.SetReadDeadline(time.Now().Add(h.heartbeat * 3))
	if err != nil {
		return err
	}

	for {
		frame, err := ws.ReadFrame(conn)
		if err != nil {
			return err
		}
		if frame.Header.OpCode == ws.OpClose {
			return errors.New("remote side close the channel")
		}
		if frame.Header.OpCode == ws.OpText {
			h.recv <- frame.Payload
		}

		if frame.Header.OpCode == ws.OpPong {
			_ = h.conn.SetReadDeadline(time.Now().Add(h.heartbeat * 3))
		}
	}
}

func (h *handler) SendText(msg string) interface{} {
	logrus.Info("send text", msg)
	return wsutil.WriteClientText(h.conn, []byte(msg))
}

func connect(addr string) (*handler, error) {
	_, err := url.Parse(addr)
	if err != nil {
		return nil, err
	}

	conn, _, _, err := ws.Dial(context.Background(), addr)
	if err != nil {
		return nil, err
	}

	h := handler{
		conn:  conn,
		close: make(chan struct{}, 1),
		recv:  make(chan []byte, 10),
	}

	go func() {
		err := h.readLoop(conn)
		if err != nil {
			logrus.Warn(err)
		}
		h.close <- struct{}{}
	}()

	return &h, nil
}

func (h *handler) heartbeatLoop() error {
	logrus.Info("heartbeat loop")
	tick := time.NewTicker(h.heartbeat)
	for range tick.C {
		logrus.Info("ping")
		if err := wsutil.WriteClientMessage(h.conn, ws.OpPing, nil); err != nil {
			return err
		}
	}
	return nil
}
