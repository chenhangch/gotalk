package container

import (
	"hash/crc32"

	"github.com/chang144/gotalk/internal/him"
	"github.com/chang144/gotalk/internal/him/wire/pkt"
)

// HashSelector 哈希选择器
type HashSelector struct {
}

// Lookup TODO： 目前的选择器是伪哈希，当service的数量不变时，固定可用，当service数量发生变化时，会导致service的哈希取值发生变化
func (h *HashSelector) Lookup(header *pkt.Header, servers []him.Service) string {
	ln := len(servers)
	code := HashCode(header.ChannelId)
	return servers[code%ln].ServiceID()
}

func HashCode(key string) int {
	hash32 := crc32.NewIEEE()
	hash32.Write([]byte(key))
	return int(hash32.Sum32())
}
