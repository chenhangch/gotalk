package container

import (
	"github.com/chang144/gotalk/internal/him"
	"github.com/chang144/gotalk/internal/him/wire/pkt"
	"hash/crc32"
)

// HashSelector 哈希选择器
type HashSelector struct {
}

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
