package yevna

import (
	"context"
	"path/filepath"

	"github.com/cockroachdb/errors"

	"github.com/tlipoca9/yevna/tracer"
)

type Context struct {
	workdir string
	silent  bool
	tracer  tracer.Tracer

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

func (c *Context) Tracer(t ...tracer.Tracer) tracer.Tracer {
	if len(t) > 1 {
		panic("too many arguments")
	}
	if len(t) == 1 {
		c.tracer = t[0]
	}
	return c.tracer
}

func (c *Context) Next(in any) (any, error) {
	c.index++
	for c.index < len(c.handlers) {
		out, err := c.handlers[c.index].Handle(c, in)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to handle %d", c.index)
		}
		c.index++
		in = out
	}
	return in, nil
}

func (c *Context) copy() *Context {
	cc := &Context{
		silent:  c.silent,
		tracer:  c.tracer,
		workdir: c.workdir,
		index:   -1,
	}
	return cc
}

func (c *Context) Run(ctx context.Context, handlers ...Handler) error {
	cc := c.copy()
	cc.ctx = ctx
	cc.handlers = handlers
	_, err := cc.Next(nil)
	return err
}

func New() *Context {
	return &Context{
		ctx:    context.Background(),
		silent: false,
		tracer: tracer.Discard,
	}
}
