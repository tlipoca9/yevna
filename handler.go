package yevna

import (
	"slices"
)

// HandlersChain defines a Handler slice.
type HandlersChain []Handler

// Copy returns a copy of the HandlersChain.
func (c HandlersChain) Copy() HandlersChain {
	return slices.Clone(c)
}

// Handler is an interface that defines a handler.
type Handler interface {
	Handle(c *Context, in any) (any, error)
}

// HandlerFunc defines a function type that implements Handler.
type HandlerFunc func(c *Context, in any) (any, error)

// Handle implements Handler.
func (f HandlerFunc) Handle(c *Context, in any) (any, error) {
	return f(c, in)
}
