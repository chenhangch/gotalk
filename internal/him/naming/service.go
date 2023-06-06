package naming

import (
	"fmt"
	"github.com/chang144/gotalk/internal/him"
)

// RegisterService Service Impl
type RegisterService struct {
	Id        string
	Name      string
	Address   string
	Port      int
	Protocol  string
	Namespace string
	Tags      []string
	Meta      map[string]string
}

// NewEntry TODO: 加上接口模式
func NewEntry(id, name, protocol string, address string, port int) him.ServiceRegistration {
	return &RegisterService{
		Id:       id,
		Name:     name,
		Address:  address,
		Port:     port,
		Protocol: protocol,
	}
}

func (s *RegisterService) ServiceName() string {
	return s.Name
}

func (s *RegisterService) PublicAddress() string {
	return s.Address
}

func (s *RegisterService) PublicPort() int {
	return s.Port
}

func (s *RegisterService) DialURL() string {
	if s.Protocol == "tcp" {
		return fmt.Sprintf("%s:%d", s.Address, s.Port)
	}
	return fmt.Sprintf("%s://%s:%d", s.Protocol, s.Address, s.Port)
}

func (s *RegisterService) GetProtocol() string {
	return s.Protocol
}

func (s *RegisterService) GetNamespace() string {
	return s.Namespace
}

func (s *RegisterService) GetTags() []string {
	return s.Tags
}

func (s *RegisterService) GetMeta() map[string]string {
	return s.Meta
}

func (s *RegisterService) String() string {
	return fmt.Sprintf("Id:%s,Name:%s,Address:%s,Port:%d,Ns:%s,Tags:%v,Meta:%v", s.Id, s.Name, s.Address, s.Port, s.Namespace, s.Tags, s.Meta)
}

func (s *RegisterService) ServiceID() string {
	return s.Id
}
