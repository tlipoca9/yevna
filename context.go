package yevna

import (
	"context"
	"path/filepath"
)

type Context struct {
	workdir string
	silent  bool

	ctx context.Context

	index    int
	handlers HandlersChain
}

func (c *Context) Context() context.Context {
	return c.ctx
}

func (c *Context) Workdir(wd ...string) string {
	if len(wd) > 1 {
		panic("too many arguments")
	}
	if len(wd) == 1 {
		path := wd[0]
		if filepath.IsLocal(wd[0]) {
			path = filepath.Join(c.workdir, path)
		}
		c.workdir = path
	}

	return c.workdir
}

func (c *Context) Silent(s ...bool) bool {
	if len(s) > 1 {
		panic("too many arguments")
	}
	if len(s) == 1 {
		c.silent = s[0]
	}
	return c.silent
}

func (c *Context) Next(in any) (any, error) {
	c.index++
	for c.index < len(c.handlers) {
		out, err := c.handlers[c.index].Handle(c, in)
		if err != nil {
			return nil, err
		}
		c.index++
		in = out
	}
	return in, nil
}

func (c *Context) copy() *Context {
	cc := &Context{
		silent:   c.silent,
		workdir:  c.workdir,
		index:    -1,
		handlers: c.handlers.Copy(),
	}
	return cc
}

func (c *Context) Use(handles ...Handler) *Context {
	c.handlers = append(c.handlers, handles...)
	return c
}

func (c *Context) Run(ctx context.Context, handlers ...Handler) error {
	cc := c.copy()
	cc.ctx = ctx
	cc.handlers = append(cc.handlers, handlers...)
	_, err := cc.Next(nil)
	return err
}

func New() *Context {
	return &Context{}
}
