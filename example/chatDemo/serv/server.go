package serv

import (
	"encoding/binary"
	"fmt"
	"github.com/gobwas/ws/wsutil"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/chang144/golunzi/errors"
	"github.com/gobwas/ws"
	"github.com/sirupsen/logrus"
)

// ServerOptions ServerOptions
type ServerOptions struct {
	writewait time.Duration //写超时时间
	readwait  time.Duration //读超时时间
}

type Server struct {
	once    sync.Once
	id      string
	address string

	options ServerOptions

	sync.Mutex
	// 会话列表
	users map[string]net.Conn
}

func NewServer(id, address string) *Server {
	return newServer(id, address)
}

func newServer(id string, address string) *Server {
	return &Server{
		id:      id,
		address: address,
		users:   make(map[string]net.Conn, 100),
		options: ServerOptions{
			writewait: time.Second * 10,
			readwait:  time.Minute * 2,
		},
	}
}

func (s *Server) Start() error {
	mux := http.NewServeMux()
	log := logrus.WithFields(logrus.Fields{
		"module": "Server",
		"listen": s.address,
		"id":     s.id,
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// step1
		conn, _, _, err := ws.UpgradeHTTP(r, w)
		if err != nil {
			conn.Close()
			return
		}

		// step2
		user := r.URL.Query().Get("user")
		if user == "" {
			conn.Close()
			return
		}

		// strp3
		old, ok := s.addUser(user, conn)
		if ok {
			_ = old.Close()
		}

		go func(user string, conn net.Conn) {
			// 读取消息
			err := s.readLoop(user, conn)
			if err != nil {
				log.Error(err)
			}
			conn.Close()
			s.delUser(user)

			log.Info("connection of %s closed", user)
		}(user, conn)
	})
	log.Infoln("started")
	return http.ListenAndServe(s.address, mux)
}

func (s *Server) addUser(user string, conn net.Conn) (net.Conn, bool) {
	s.Lock()
	defer s.Unlock()
	old, ok := s.users[user]
	s.users[user] = conn
	return old, ok
}

func (s *Server) delUser(user string) {
	s.Lock()
	defer s.Unlock()
	delete(s.users, user)
}

func (s *Server) Shutdown() {
	s.once.Do(func() {
		s.Lock()
		defer s.Unlock()
		for _, conn := range s.users {
			conn.Close()
		}
	})
}

// readLoop 读取客户端发送的数据
func (s *Server) readLoop(user string, conn net.Conn) error {
	for {
		_ = conn.SetReadDeadline(time.Now().Add(s.options.readwait))

		// 从TCP缓冲中读取一帧的消息
		frame, err := ws.ReadFrame(conn)
		if err != nil {
			return err
		}
		if frame.Header.OpCode == ws.OpClose {
			return errors.New("remote side close the conn")
		}

		if frame.Header.Masked {
			// 使用Mask解码数据包
			ws.Cipher(frame.Payload, frame.Header.Mask, 0)
		}

		if frame.Header.OpCode == ws.OpText {
			go s.handle(user, string(frame.Payload))
		} else if frame.Header.OpCode == ws.OpBinary {
			go s.handleBinary(user, frame.Payload)
		}

		if frame.Header.OpCode == ws.OpPing {
			_ = wsutil.WriteServerMessage(conn, ws.OpPong, nil)
			continue
		}

	}
}

// handler 广播消息
func (s *Server) handle(user string, message string) {
	logrus.Info("recv message %s", message, user)
	s.Lock()
	defer s.Unlock()

	broadcast := fmt.Sprintf("%s --- FORM %s", message, user)
	for u, conn := range s.users {
		// 消息不发生给自己
		if u == user {
			continue
		}
		logrus.Infof("send to %s : %s", u, broadcast)
		err := s.writeText(conn, broadcast)
		if err != nil {
			logrus.Errorf("write to %s failed, error :%v", user, err)
		}
	}
}

func (s *Server) writeText(conn net.Conn, message string) error {
	f := ws.NewTextFrame([]byte(message))
	return ws.WriteFrame(conn, f)
}

// 服务端处理ping消息
const (
	CommandPing = 100 + iota
	CommandPong
)

func (s *Server) handleBinary(user string, message []byte) error {
	logrus.Info("recv message handleBinary %s from %s", message, user)
	s.Lock()
	defer s.Unlock()
	// handle ping request
	i := 0
	command := binary.BigEndian.Uint16(message[i : i+2])
	i += 2
	payloadLen := binary.BigEndian.Uint32(message[i : i+4])
	logrus.Infof("command: %v payloadLen: %v", command, payloadLen)
	if command == CommandPing {
		u := s.users[user]
		// return pong
		err := wsutil.WriteServerBinary(u, []byte{0, CommandPong, 0, 0, 0, 0})
		if err != nil {
			logrus.Errorf("write to %s failed, error: %v", user, err)
		}
	}
	return nil
}
