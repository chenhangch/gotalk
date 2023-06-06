package him

import (
	"sync"
	"sync/atomic"
)

type Event struct {
	fired int32
	c     chan struct{}
	o     sync.Once
}

// Fire causes e to complete.  It is safe to call multiple times, and
// concurrently.  It returns true iff this call to Fire caused the signaling
// channel returned by Done to close.
// 用于触发事件，它通过sync.Once确保事件只会被触发一次，
// 使用atomic.StoreInt32()原子操作将fired标志位置为1，表示事件已经触发。
// 同时，它还会关闭通道c，以通知等待该事件的goroutine
func (e *Event) Fire() bool {
	ret := false
	e.o.Do(func() {
		atomic.StoreInt32(&e.fired, 1)
		close(e.c)
		ret = true
	})
	return ret
}

func (e *Event) Done() <-chan struct{} {
	return e.c
}

func (e *Event) HasFired() bool {
	return atomic.LoadInt32(&e.fired) == 1
}

func NewEvent() *Event {
	return &Event{
		c: make(chan struct{}),
	}
}
