package tcp

import (
	"context"
	"errors"
	"fmt"
	"github.com/chang144/gotalk/internal/him"
	"github.com/klintcheng/kim/logger"
	"github.com/segmentio/ksuid"
	"net"
	"sync"
	"time"
)

type ServerOptions struct {
	loginWait time.Duration
	writeWait time.Duration
	readWait  time.Duration
}

// Server is a tcp implement of him.Server
type Server struct {
	listen string

	him.ServiceRegistration
	him.ChannelMap
	him.Acceptor
	him.MessageListener
	him.StateListener

	once    sync.Once
	options ServerOptions
	quit    *him.Event
}

func NewServer(listen string, service him.ServiceRegistration) him.Server {
	return &Server{
		listen:              listen,
		ServiceRegistration: service,
		ChannelMap:          him.NewChannelMap(100),

		quit: him.NewEvent(),
		options: ServerOptions{
			loginWait: him.DefaultLoginWait,
			writeWait: him.DefaultWriteWait,
			readWait:  time.Second * 10,
		},
	}
}

func (s *Server) Start() error {
	log := logger.WithFields(logger.Fields{
		"module": "tcp.chatServer",
		"listen": s.listen,
		"id":     s.ServiceID(),
	})

	if s.StateListener == nil {
		return fmt.Errorf("StateListener is nil")
	}
	if s.Acceptor == nil {
		s.Acceptor = new(defaultAcceptor)
	}
	// step 1
	lst, err := net.Listen("tcp", s.listen)
	if err != nil {
		return err
	}
	log.Info("starting tcp chatServer")
	for {
		//step 2
		rawconn, err := lst.Accept()
		if err != nil {
			rawconn.Close()
			log.Warn(err)
			continue
		}
		go func(rawconn net.Conn) {
			conn := NewConn(rawconn)
			// step 3
			id, err := s.Accept(conn, s.options.loginWait)
			if err != nil {
				_ = conn.WriteFrame(him.OpClose, []byte(err.Error()))
				conn.Close()
				return
			}
			if _, ok := s.Get(id); ok {
				log.Warn("channel %s existed", id)
				_ = conn.WriteFrame(him.OpClose, []byte("channelId is exists"))
				conn.Close()
				return
			}
			// step 4
			channel := him.NewChannel(id, conn)
			channel.SetReadWait(s.options.readWait)
			channel.SetWriteWait(s.options.writeWait)
			s.Add(channel)

			//step 5
			err = channel.ReadLoop(s.MessageListener)
			if err != nil {
				log.Info(err)
			}
			// step 6
			s.Remove(channel.ID())
			_ = s.Disconnect(channel.ID())
			channel.Close()
		}(rawconn)
	}
}

// Push string channelID
// []byte data
func (s *Server) Push(id string, data []byte) error {
	ch, ok := s.ChannelMap.Get(id)
	if !ok {
		return errors.New("channel no found")
	}
	return ch.Push(data)
}

func (s *Server) Shutdown(ctx context.Context) error {
	logger.WithFields(logger.Fields{
		"module": "tcp.chatServer",
		"id":     s.ServiceID(),
	})
	s.once.Do(func() {
		defer func() {
			logger.Infoln("shutdown")
		}()
		channels := s.ChannelMap.All()
		for _, channel := range channels {
			channel.Close()

			select {
			case <-ctx.Done():
				return
			default:
				continue
			}
		}
	})
	return nil
}

func (s *Server) SetAcceptor(acceptor him.Acceptor) {
	s.Acceptor = acceptor
}

func (s *Server) SetMessageListener(listener him.MessageListener) {
	s.MessageListener = listener
}

func (s *Server) SetStateListener(listener him.StateListener) {
	s.StateListener = listener
}

func (s *Server) SetReadWait(readWait time.Duration) {
	s.options.readWait = readWait
}

func (s *Server) SetChannelMap(channelMap him.ChannelMap) {
	s.ChannelMap = channelMap
}

type defaultAcceptor struct {
}

func (d defaultAcceptor) Accept(conn him.Conn, duration time.Duration) (string, error) {
	return ksuid.New().String(), nil
}
