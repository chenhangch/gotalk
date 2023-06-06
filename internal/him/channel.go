package him

import (
	"errors"
	"sync"
	"time"

	"github.com/klintcheng/kim/logger"
)

// Channel 上层通用逻辑的封装
type Channel interface {
	Conn
	Agent
	Close() error //overwrite net.Conn.Close()
	ReadLoop(MessageListener) error
	SetWriteWait(time.Duration)
	SetReadWait(time.Duration)
}

type channelImpl struct {
	sync.Mutex
	id string
	Conn
	writeChan chan []byte
	once      sync.Once
	writeWait time.Duration
	readWait  time.Duration
	closed    *Event
}

func NewChannel(id string, conn Conn) Channel {
	log := logger.WithFields(logger.Fields{
		"module": "tcp_channel",
		"id":     id,
	})
	ch := &channelImpl{
		id:        id,
		Conn:      conn,
		writeChan: make(chan []byte, 5),
		writeWait: DefaultWriteWait,
		readWait:  DefaultReadWait,
		closed:    NewEvent(),
	}
	go func() {
		err := ch.writeLoop()
		if err != nil {
			log.Info(err)
		}
	}()

	return ch
}

func (ch *channelImpl) writeLoop() error {
	for {
		select {
		case payload := <-ch.writeChan:
			err := ch.WriteFrame(OpBinary, payload)
			if err != nil {
				return err
			}
			chanlen := len(ch.writeChan)
			for i := 0; i < chanlen; i++ {
				payload = <-ch.writeChan
				err := ch.WriteFrame(OpBinary, payload)
				if err != nil {
					return err
				}
			}
			err = ch.Conn.Flush()
			if err != nil {
				return err
			}
		case <-ch.closed.Done():
			return nil
		}
	}
}

// ReadLoop read loop from channel
// 是一个阻塞的方法，把消息的读取和心跳处理的逻辑封装在一起
func (ch *channelImpl) ReadLoop(lst MessageListener) error {
	ch.Lock()
	defer ch.Unlock()
	log := logger.WithFields(logger.Fields{
		"struct": "ChannelImpl",
		"func":   "ReadLoop",
		"id":     ch.id,
	})
	for {
		_ = ch.SetReadDeadline(time.Now().Add(ch.readWait))

		frame, err := ch.ReadFrame()
		if err != nil {
			return err
		}
		if frame.GetOpCode() == OpClose {
			return errors.New("remote side close the channel")
		}
		if frame.GetOpCode() == OpPing {
			log.Trace("recv a ping: resp with a pong")
			_ = ch.WriteFrame(OpPong, nil)
			continue
		}
		payload := frame.GetPayload()
		if len(payload) == 0 {
			continue
		}
		go lst.Receive(ch, payload)
	}

}

// Push message to channel
// 通过管道writeChan，发送给一个独立的goruntine中的writeLoop()执行,使得Push变成了一个线程安全方法
func (ch *channelImpl) Push(payload []byte) error {
	if ch.closed.HasFired() {
		return errors.New("channel")
	}
	// 异步写
	ch.writeChan <- payload
	return nil
}

// WriteFrame overwrite him.Conn.WriteFrame
// 增加了重置写超时的逻辑
func (ch *channelImpl) WriteFrame(code OpCode, payload []byte) error {
	err := ch.Conn.SetWriteDeadline(time.Now().Add(ch.writeWait))
	if err != nil {
		return err
	}
	return ch.Conn.WriteFrame(code, payload)
}

func (ch *channelImpl) ID() string {
	return ch.id
}

func (ch *channelImpl) SetWriteWait(writeWait time.Duration) {
	if writeWait == 0 {
		return
	}
	ch.writeWait = writeWait
}

func (ch *channelImpl) SetReadWait(readwait time.Duration) {
	if readwait == 0 {
		return
	}
	ch.readWait = readwait
}

// ChannelMap ChannelMap
type ChannelMap interface {
	Add(Channel)
	Remove(id string)
	Get(id string) (Channel, bool)
	All() []Channel
}

type ChannelMapImpl struct {
	channels *sync.Map
}

// Add channel to channelMap
func (c *ChannelMapImpl) Add(channel Channel) {
	if channel.ID() == "" {
		logger.WithFields(logger.Fields{
			"module": "ChannelsImpl",
		}).Error("channel id is required")
	}

	c.channels.Store(channel.ID(), channel)
}

func (c *ChannelMapImpl) Remove(id string) {
	c.channels.Delete(id)
}

func (c *ChannelMapImpl) Get(id string) (Channel, bool) {
	if id == "" {
		logger.WithFields(logger.Fields{
			"module": "ChannelsImpl",
		}).Error("Channel id is required")
	}

	val, ok := c.channels.Load(id)
	if !ok {
		return nil, false
	}
	return val.(Channel), true
}

func (c *ChannelMapImpl) All() []Channel {
	arr := make([]Channel, 0)
	c.channels.Range(func(key, val any) bool {
		arr = append(arr, val.(Channel))
		return true
	})
	return arr
}

func NewChannelMap(num int) ChannelMap {
	return &ChannelMapImpl{
		channels: new(sync.Map),
	}
}
