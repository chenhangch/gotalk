package container

import (
	"github.com/chang144/gotalk/internal/him"
	"github.com/chang144/gotalk/internal/him/wire/pkt"
)

// Selector is used to select a Service
type Selector interface {
	Lookup(*pkt.Header, []him.Service) string
}
