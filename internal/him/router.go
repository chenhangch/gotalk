package him

import "sync"

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

func BuildContext() Context {
	return &ContextImpl{}
}

type FuncTree struct {
	nodes map[string]HandlersChain
}

// NewTree New Tree
func NewTree() *FuncTree {
	return &FuncTree{nodes: make(map[string]HandlersChain, 10)}
}
