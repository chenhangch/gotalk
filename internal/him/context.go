package him

import (
	"github.com/chang144/gotalk/internal/him/wire/pkt"
	"github.com/klintcheng/kim/logger"
	"google.golang.org/protobuf/proto"
)

type Context interface {
	Dispatcher

	Header() *pkt.Header
	ReadBody(val proto.Message) error
	Session() Session
	RespWithError(status pkt.Status, err error) error
	Resp(status pkt.Status, body proto.Message) error
	Dispatch(body proto.Message, revs ...*Location) error
	Next()
}

// HandlerFunc defines the handler used
type HandlerFunc func(ctx Context)

// HandlersChain HandlersChain
type HandlersChain []HandlerFunc

// ContextImpl is the most important part of him
type ContextImpl struct {
	Dispatcher

	requestPkt *pkt.LogicPkt

	session Session
}

func (c *ContextImpl) Header() *pkt.Header {
	//TODO implement me
	panic("implement me")
}

func (c *ContextImpl) ReadBody(val proto.Message) error {
	//TODO implement me
	panic("implement me")
}

func (c *ContextImpl) Session() Session {
	//TODO implement me
	panic("implement me")
}

func (c *ContextImpl) RespWithError(status pkt.Status, err error) error {
	//TODO implement me
	panic("implement me")
}

func (c *ContextImpl) Resp(status pkt.Status, body proto.Message) error {
	//TODO implement me
	panic("implement me")
}

func (c *ContextImpl) Next() {
	//TODO implement me
	panic("implement me")
}

func (c *ContextImpl) Dispatch(body proto.Message, recs ...*Location) error {
	if len(recs) == 0 {
		return nil
	}

	logicPkt := pkt.NewLogicPkt(&c.requestPkt.Header)
	logicPkt.Flag = pkt.Flag_Push
	logicPkt.WriteBody(body)

	group := make(map[string][]string)
	for _, recv := range recs {
		if recv.ChannelId == c.session.GetChannelId() {
			continue
		}
		if _, ok := group[recv.GateId]; !ok {
			group[recv.GateId] = make([]string, 0)
		}
		group[recv.GateId] = append(group[recv.GateId], recv.ChannelId)
	}

	for gateway, ids := range group {
		err := c.Push(gateway, ids, logicPkt)
		if err != nil {
			logger.Error(err)
		}
		return err
	}

	return nil
}

var _ Context = (*ContextImpl)(nil)
