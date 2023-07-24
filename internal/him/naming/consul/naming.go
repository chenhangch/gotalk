package consul

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/chang144/gotalk/internal/him"
	"github.com/chang144/gotalk/internal/him/naming"
	"github.com/hashicorp/consul/api"
	"github.com/klintcheng/kim/logger"
)

const (
	KeyProtocol  = "protocol"
	KeyHealthURL = "health_url"
)

const (
	HealthPassing     = "passing"
	HealthWarning     = "warning"
	HealthCritical    = "critical"
	HealthMaintenance = "maintenance"
)

type Watch struct {
	Service   string
	Callback  func([]him.ServiceRegistration)
	WaitIndex uint64
	Quit      chan struct{}
}

type Naming struct {
	sync.RWMutex
	cli    *api.Client
	watchs map[string]*Watch
}

func (n *Naming) Find(serviceName string, tags ...string) ([]him.ServiceRegistration, error) {
	services, _, err := n.load(serviceName, 0, tags...)
	if err != nil {
		return nil, err
	}
	return services, nil
}

// load 用于服务发现
// waitIndex: 阻塞查询, 传0表示不阻塞
func (n *Naming) load(name string, waitIndex uint64, tags ...string) ([]him.ServiceRegistration, *api.QueryMeta, error) {
	opts := &api.QueryOptions{
		UseCache:  true,
		MaxAge:    time.Minute,
		WaitIndex: waitIndex,
	}
	catalogServices, meta, err := n.cli.Catalog().ServiceMultipleTags(name, tags, opts)
	if err != nil {
		return nil, meta, err
	}

	service := make([]him.ServiceRegistration, 0, len(catalogServices))
	for _, s := range catalogServices {
		if s.Checks.AggregatedStatus() != api.HealthPassing {
			logger.Debugf("load service: id:%s name:%s %s:%d Status:%s", s.ServiceID, s.ServiceName, s.ServiceAddress, s.ServicePort, s.Checks.AggregatedStatus())
			continue
		}
		service = append(service, &naming.RegisterService{
			Id:       s.ServiceID,
			Name:     s.ServiceName,
			Address:  s.ServiceAddress,
			Port:     s.ServicePort,
			Protocol: s.ServiceMeta[KeyProtocol],
			Tags:     s.ServiceTags,
			Meta:     s.ServiceMeta,
		})
	}
	logger.Debugf("load service: %v, meta:%v", service, meta)
	return service, meta, nil
}

func (n *Naming) Register(s him.ServiceRegistration) error {
	reg := &api.AgentServiceRegistration{
		ID:      s.ServiceID(),
		Name:    s.ServiceName(),
		Address: s.PublicAddress(),
		Port:    s.PublicPort(),
		Tags:    s.GetTags(),
		Meta:    s.GetMeta(),
	}
	if reg.Meta == nil {
		reg.Meta = make(map[string]string)
	}
	// 将协议类型加到Meta，在服务消费方可以知道服务提供的接入协议
	reg.Meta[KeyProtocol] = s.GetProtocol()

	// consul 健康检查
	healthURL := s.GetMeta()[KeyHealthURL]
	if healthURL != "" {
		// 健康回调函数
		check := new(api.AgentServiceCheck)
		check.CheckID = fmt.Sprintf("%s_normal", s.ServiceID())
		check.HTTP = healthURL
		check.Timeout = "1s"
		check.Interval = "10s"
		check.DeregisterCriticalServiceAfter = "20s"
		reg.Check = check
	}
	err := n.cli.Agent().ServiceRegister(reg)
	return err
}

func (n *Naming) Deregister(serviceID string) error {
	return n.cli.Agent().ServiceDeregister(serviceID)
}

func (n *Naming) Subscribe(serviceName string, callback func(services []him.ServiceRegistration)) error {
	n.Lock()
	defer n.Unlock()
	if _, ok := n.watchs[serviceName]; ok {
		return errors.New("serviceName has already been registered")
	}
	w := &Watch{
		Service:  serviceName,
		Callback: callback,
		Quit:     make(chan struct{}, 1),
	}

	go n.watch(w)
	return nil
}

func (n *Naming) Unsubscribe(serviceName string) error {
	n.Lock()
	defer n.Unlock()
	wh, ok := n.watchs[serviceName]
	delete(n.watchs, serviceName)
	if ok {
		close(wh.Quit)
	}
	return nil
}

func (n *Naming) watch(wh *Watch) {
	stopped := false

	var doWatch = func(service string, callback func([]him.ServiceRegistration)) {
		services, meta, err := n.load(service, wh.WaitIndex)
		if err != nil {
			return
		}
		select {
		case <-wh.Quit:
			stopped = true
			logger.Infof("watch %s stopped", wh.Service)
			return
		default:
		}

		wh.WaitIndex = meta.LastIndex
		if callback != nil {
			callback(services)
		}
	}

	// build WaitIndex
	doWatch(wh.Service, nil)
	for !stopped {
		doWatch(wh.Service, wh.Callback)
	}
}

func NewNaming(consulUrl string) (naming.Naming, error) {
	conf := api.DefaultConfig()
	conf.Address = consulUrl
	cli, err := api.NewClient(conf)
	if err != nil {
		return nil, err
	}
	n := &Naming{
		cli:    cli,
		watchs: make(map[string]*Watch, 1),
	}

	return n, nil
}
