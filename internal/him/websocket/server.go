package websocket

import (
	"context"
	"errors"
	"fmt"
	"github.com/segmentio/ksuid"
	"net/http"
	"sync"
	"time"

	"github.com/chang144/gotalk/internal/him"
	"github.com/gobwas/ws"
	"github.com/klintcheng/kim/logger"
)

type ServerOptions struct {
	loginWait time.Duration
	readWait  time.Duration
	writeWait time.Duration
}

// Server is a websocket implement of the Server interface
type Server struct {
	listen string

	him.ServiceRegistration
	him.ChannelMap
	him.Acceptor
	him.MessageListener
	him.StateListener

	once    sync.Once
	options ServerOptions
}

func NewServer(listen string, service him.ServiceRegistration) him.Server {
	return &Server{
		listen:              listen,
		ServiceRegistration: service,
		options: ServerOptions{
			loginWait: him.DefaultLoginWait,
			readWait:  him.DefaultReadWait,
			writeWait: him.DefaultWriteWait,
		},
	}
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

func (s *Server) Start() error {
	mux := http.NewServeMux()
	log := logger.WithFields(logger.Fields{
		"module": "ws.chatServer",
		"listen": s.listen,
		"id":     s.ServiceID(),
	})

	if s.Acceptor == nil {
		s.Acceptor = new(defaultAcceptor)
	}
	if s.StateListener == nil {
		return fmt.Errorf("StateListener is nil")
	}

	// 连接管理器
	if s.ChannelMap == nil {
		s.ChannelMap = him.NewChannelMap(100)
	}

	mux.HandleFunc("/im", func(w http.ResponseWriter, r *http.Request) {
		// 握手升级成websocket长连接
		rawconn, _, _, err := ws.UpgradeHTTP(r, w)
		if err != nil {
			resp(w, http.StatusBadRequest, err.Error())
		}

		// 包装conn
		conn := NewConn(rawconn)

		// 鉴权得到该管道id
		id, err := s.Accept(conn, s.options.loginWait)
		if err != nil {
			_ = conn.WriteFrame(him.OpClose, []byte(err.Error()))
			conn.Close()
			return
		}

		if _, ok := s.Get(id); !ok {
			log.Warnf("channel %s existed", id)
			_ = conn.WriteFrame(him.OpClose, []byte("channelId is repeated"))
			conn.Close()
			return
		}

		// 创建channel 并添加到channelMap
		channel := him.NewChannel(id, conn)
		channel.SetReadWait(s.options.readWait)
		channel.SetWriteWait(s.options.writeWait)
		s.Add(channel)

		go func(ch him.Channel) {
			// 5
			err := ch.ReadLoop(s.MessageListener)
			if err != nil {
				log.Info(err)
			}
			// 6
			s.Remove(ch.ID())
			err = s.Disconnect(ch.ID())
			if err != nil {
				log.Warn(err)
			}
			ch.Close()
		}(channel)
	})

	log.Infoln("started chatServer")
	return http.ListenAndServe(s.listen, mux)

}

func (s *Server) Push(id string, data []byte) error {
	ch, ok := s.ChannelMap.Get(id)
	if !ok {
		return errors.New("channel to found")
	}
	return ch.Push(data)
}

func (s *Server) Shutdown(ctx context.Context) error {
	log := logger.WithFields(logger.Fields{
		"module": "ws.chatServer",
		"id":     s.ServiceID(),
	})
	s.once.Do(func() {
		defer func() {
			log.Infoln("shutdown")
		}()
		// close channels
		channelMap := s.ChannelMap.All()
		for _, channel := range channelMap {
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

func resp(w http.ResponseWriter, code int, body string) {
	w.WriteHeader(code)
	if body != "" {
		_, err := w.Write([]byte(body))
		if err != nil {
			return
		}
	}
	logger.Warnf("response with backend:%d %s", code, body)
}

type defaultAcceptor struct {
}

func (s *defaultAcceptor) Accept(conn him.Conn, timeout time.Duration) (string, error) {
	return ksuid.New().String(), nil
}
