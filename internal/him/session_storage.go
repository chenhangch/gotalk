package him

import (
	"errors"
	"github.com/chang144/gotalk/internal/him/wire/pkt"
)

// ErrNil
var ErrSessionNil = errors.New("err:session nil")

// SessionStorage 定义会话存储，提供基于保存、删除、查找会话的功能
type SessionStorage interface {
	// Add a session
	Add(session *pkt.Session) error
	// Delete a session
	Delete(account string, channelId string) error
	// Get session by channelId
	Get(channelId string) (*pkt.Session, error)
	// GetLocations Get Locations by accounts
	GetLocations(account ...string) ([]*Location, error)
	// GetLocation Get Location by account and device
	GetLocation(account string, device string) (*Location, error)
}
