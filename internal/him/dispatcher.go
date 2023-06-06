package him

import "github.com/chang144/gotalk/internal/him/wire/pkt"

// Dispatcher 向网关中的channels两个连接推送一条消息LogicPkt
// 这个能力由容器提供
type Dispatcher interface {
	Push(gateway string, channels []string, p *pkt.LogicPkt) error
}
