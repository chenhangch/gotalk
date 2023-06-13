package him

import (
	"fmt"
	"github.com/chang144/gotalk/internal/him/wire/pkt"
	"sync"
)

type Router struct {
	// 中间件
	middleware []HandlerFunc
	// 注册的监听器列表
	handlers *FuncTree
	// 对象池
	pool sync.Pool
}

func NewRouter() *Router {
	r := &Router{
		middleware: make([]HandlerFunc, 0),
		handlers:   NewTree(),
		pool:       sync.Pool{},
	}
	r.pool.New = func() any {
		return BuildContext()
	}
	return r
}

// AddHandles 添加handlers
func (r *Router) AddHandles(command string, handlers ...HandlerFunc) {
	r.handlers.Add(command, handlers...)
}

func (r *Router) Serve(pkt *pkt.LogicPkt, dispatcher Dispatcher, cache SessionStorage, session Session) error {
	if dispatcher == nil {
		return fmt.Errorf("no dispatcher")
	}
	if cache == nil {
		return fmt.Errorf("no cache")
	}
	ctx := r.pool.Get().(*ContextImpl)
	ctx.reset()
	ctx.requestPkt = pkt
	ctx.SessionStorage = cache
	ctx.session = session

	r.serveContext(ctx)
	r.pool.Put(ctx)

	return nil
}

func (r *Router) serveContext(ctx *ContextImpl) {
	chain, ok := r.handlers.Get(ctx.Header().Command)
	if !ok {
		ctx.handlers = []HandlerFunc{handleNoFound}
		ctx.Next()
		return
	}
	ctx.handlers = chain
	ctx.Next()
}

func handleNoFound(ctx Context) {
	_ = ctx.Resp(pkt.Status_NotImplemented, &pkt.ErrorResp{Message: "NotImplemented"})
}

type FuncTree struct {
	nodes map[string]HandlersChain
}

// NewTree New Tree
func NewTree() *FuncTree {
	return &FuncTree{nodes: make(map[string]HandlersChain, 10)}
}

func (t *FuncTree) Add(path string, handlers ...HandlerFunc) {
	t.nodes[path] = append(t.nodes[path], handlers...)
}

func (t *FuncTree) Get(path string) (HandlersChain, bool) {
	f, ok := t.nodes[path]
	return f, ok
}
