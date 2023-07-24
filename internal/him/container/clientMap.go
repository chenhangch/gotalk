package container

import (
	"sync"

	"github.com/chang144/gotalk/internal/him"
	"github.com/klintcheng/kim/logger"
)

type ClientMap interface {
	Add(client him.Client)
	Remove(clientId string)
	Get(clientId string) (client him.Client, ok bool)

	Services(kvs ...string) []him.Service
}

type ClientMapImpl struct {
	client *sync.Map
}

func (ch *ClientMapImpl) Add(client him.Client) {
	if client.ServiceID() == "" {
		logger.WithFields(logger.Fields{
			"module": "ClientsImpl",
		}).Error("client id is required")
	}
	ch.client.Store(client.ServiceID(), client)
}

func (ch *ClientMapImpl) Remove(clientId string) {
	ch.client.Delete(clientId)
}

func (ch *ClientMapImpl) Get(clientId string) (client him.Client, ok bool) {
	if clientId == "" {
		logger.WithFields(logger.Fields{
			"module": "ClientsImpl",
		}).Error("client id is required")
	}

	val, ok := ch.client.Load(clientId)
	if !ok {
		return nil, false
	}
	return val.(him.Client), true
}

// Services 返回服务列表
// TODO: 优化kvs传参
func (ch *ClientMapImpl) Services(kvs ...string) []him.Service {
	kvLen := len(kvs)
	if kvLen != 0 && kvLen != 2 {
		return nil
	}
	serviceArr := make([]him.Service, 0)
	ch.client.Range(func(key, value any) bool {
		ser := value.(him.Service)
		if kvLen > 0 && ser.GetMeta()[kvs[0]] != kvs[1] {
			serviceArr = append(serviceArr, ser)
		}
		return true
	})
	return serviceArr
}

func NewClientMap(num int) ClientMap {
	return &ClientMapImpl{client: new(sync.Map)}
}
