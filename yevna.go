package yevna

import (
	"context"
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
	c := New().Use(
		ErrorHandler(),
		Recover(),
	)
	SetDefault(c)
}

func Run(ctx context.Context, handlers ...Handler) error {
	return Default().Run(ctx, handlers...)
}
