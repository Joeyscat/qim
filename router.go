package qim

import (
	"errors"
	"sync"

	"github.com/joeyscat/qim/wire/pkt"
)

var ErrSessionLost = errors.New("err:session lost")

type Router struct {
	middleware []HandlerFunc
	handlers   *FuncTree
	pool       sync.Pool
}

func NewRouter() *Router {
	r := &Router{
		handlers:   NewTree(),
		middleware: make([]HandlerFunc, 0),
	}
	r.pool.New = func() any {
		return BuildContext()
	}
	return r
}

func (r *Router) Use(handlers ...HandlerFunc) {
	r.middleware = append(r.middleware, handlers...)
}

// Handle register a command handler
func (r *Router) Handle(command string, handlers ...HandlerFunc) {
	r.handlers.Add(command, r.middleware...)
	r.handlers.Add(command, handlers...)
}

func (r *Router) Serve(packet *pkt.LogicPkt, dispacther Dispatcher,
	cache SessionStorage, session Session) error {
	if dispacther == nil {
		return errors.New("dispacther is nil")
	}
	if cache == nil {
		return errors.New("cache is nil")
	}

	ctx := r.pool.Get().(*ContextImpl)
	ctx.reset()
	ctx.request = packet
	ctx.Dispatcher = dispacther
	ctx.SessionStorage = cache
	ctx.session = session

	r.serveContext(ctx)
	r.pool.Put(ctx)

	return nil
}

func (r *Router) serveContext(ctx *ContextImpl) {
	chain, ok := r.handlers.Get(ctx.request.Command)
	if !ok {
		ctx.handlers = []HandlerFunc{handleNotFound}
		ctx.Next()
		return
	}
	ctx.handlers = chain
	ctx.Next()
}

func handleNotFound(ctx Context) {
	_ = ctx.Resp(pkt.Status_NotImplemented, &pkt.ErrorResp{Message: "NotImplemented"})
}

// FuncTree is a tree structure
type FuncTree struct {
	nodes map[string]HandlersChain
}

func NewTree() *FuncTree {
	return &FuncTree{
		nodes: make(map[string]HandlersChain, 10),
	}
}

func (t *FuncTree) Add(path string, handlers ...HandlerFunc) {
	if t.nodes[path] == nil {
		t.nodes[path] = HandlersChain{}
	}

	t.nodes[path] = append(t.nodes[path], handlers...)
}

func (t *FuncTree) Get(path string) (HandlersChain, bool) {
	f, ok := t.nodes[path]
	return f, ok
}
