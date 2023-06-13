package him

import (
	"bytes"
	"errors"
	"github.com/chang144/gotalk/internal/him/wire/endian"
)

// Location 用户位置，
type Location struct {
	// 通道Id
	ChannelId string
	// 网关Id
	GateId string
}

func (loc *Location) Bytes() []byte {
	if loc == nil {
		return []byte{}
	}
	buf := new(bytes.Buffer)
	_ = endian.WriteShortBytes(buf, []byte(loc.ChannelId))
	_ = endian.WriteShortBytes(buf, []byte(loc.GateId))
	return buf.Bytes()
}

func (loc *Location) Unmarshal(data []byte) (err error) {
	if len(data) == 0 {
		return errors.New("data is empty")
	}
	buf := bytes.NewBuffer(data)
	loc.ChannelId, err = endian.ReadShortString(buf)
	if err != nil {
		return
	}
	loc.GateId, err = endian.ReadShortString(buf)
	if err != nil {
		return
	}
	return
}
