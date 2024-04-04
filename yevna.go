package yevna

import (
	"context"
	"os"
	"sync/atomic"

	"github.com/tlipoca9/yevna/tracer"
)

var defaultContext atomic.Pointer[Context]

func Default() *Context {
	return defaultContext.Load()
}

func SetDefault(c *Context) {
	defaultContext.Store(c)
}

func init() {
	c := New().Use(Recover())
	c.Tracer(tracer.NewExecTracer(os.Stderr))
	SetDefault(c)
}

func Run(ctx context.Context, handlers ...Handler) error {
	return Default().Run(ctx, handlers...)
}
