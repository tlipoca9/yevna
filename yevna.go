package yevna

import (
	"context"
	"github.com/tlipoca9/yevna/tracer"
	"os"
	"sync/atomic"
)

var defaultContext atomic.Pointer[Context]

func Default() *Context {
	return defaultContext.Load()
}

func SetDefault(c *Context) {
	defaultContext.Store(c)
}

func init() {
	c := New()
	c.Tracer(tracer.NewExecTracer(os.Stderr))
	SetDefault(c)
}

func Run(ctx context.Context, handlers ...Handler) error {
	return Default().Run(ctx, handlers...)
}
