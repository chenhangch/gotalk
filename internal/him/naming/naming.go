package naming

import (
	"errors"
	"github.com/chang144/gotalk/internal/him"
)

var ErrNotFound = errors.New("service no found")

type Naming interface {
	// Find load all servers nodes
	Find(serviceName string, tag ...string) ([]him.ServiceRegistration, error)
	//Remove(serviceName, serviceID string) error
	// Get(namespace string, id string) (ServiceRegistration, error)
	Register(him.ServiceRegistration) error
	Deregister(serviceID string) error

	// Subscribe 订阅服务变更通知
	Subscribe(serviceName string, callback func(services []him.ServiceRegistration)) error
	// Unsubscribe 取消订阅服务变更通知
	Unsubscribe(serviceName string) error
}
