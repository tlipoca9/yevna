package yevna

import (
	"bytes"
	"context"
	"fmt"
	"github.com/cockroachdb/errors"
	"github.com/tlipoca9/yevna/tracer"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sync/atomic"
)

var defaultHandlerContext atomic.Pointer[HandlerContext]

func DefaultHandlerContext() *HandlerContext {
	return defaultHandlerContext.Load()
}

func SetDefaultHandlerContext(c *HandlerContext) {
	defaultHandlerContext.Store(c)
}

func init() {
	SetDefaultHandlerContext(NewHandlerContext())
}

// HandlersChain defines a Handler slice
type HandlersChain []Handler

type HandlerContext struct {
	workdir string
	silent  bool
	tracer  tracer.Tracer

	ctx context.Context

	index    int
	handlers HandlersChain
}

func (c *HandlerContext) Context() context.Context {
	return c.ctx
}

func (c *HandlerContext) Workdir(wd ...string) string {
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

func (c *HandlerContext) Silent(s ...bool) bool {
	if len(s) > 1 {
		panic("too many arguments")
	}
	if len(s) == 1 {
		c.silent = s[0]
	}
	return c.silent
}

func (c *HandlerContext) Tracer(t ...tracer.Tracer) tracer.Tracer {
	if len(t) > 1 {
		panic("too many arguments")
	}
	if len(t) == 1 {
		c.tracer = t[0]
	}
	return c.tracer
}

func (c *HandlerContext) Next(in any) (any, error) {
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

func (c *HandlerContext) copy() *HandlerContext {
	cc := &HandlerContext{
		silent:  c.silent,
		tracer:  c.tracer,
		workdir: c.workdir,
		index:   -1,
	}
	return cc
}

func (c *HandlerContext) Run(ctx context.Context, handlers ...Handler) error {
	cc := c.copy()
	cc.ctx = ctx
	cc.handlers = handlers
	_, err := cc.Next(nil)
	return err
}

func NewHandlerContext() *HandlerContext {
	return &HandlerContext{
		ctx:    context.Background(),
		silent: false,
		tracer: tracer.Discard,
	}
}

type Handler interface {
	Handle(c *HandlerContext, in any) (any, error)
}

type HandlerFunc func(c *HandlerContext, in any) (any, error)

func (f HandlerFunc) Handle(c *HandlerContext, in any) (any, error) {
	return f(c, in)
}

func Run(ctx context.Context, handlers ...Handler) error {
	return DefaultHandlerContext().Run(ctx, handlers...)
}

func Echo(r io.Reader) Handler {
	var (
		name = "echo"
		args []string
	)

	if r == os.Stdin {
		args = append(args, "<stdin>")
	} else {
		args = append(args, fmt.Sprintf("<%T>", r))
	}

	return HandlerFunc(func(c *HandlerContext, in any) (any, error) {
		c.Tracer().Trace(name, args...)
		return r, nil
	})
}

func Cd(path string) Handler {
	name := "cd"
	args := []string{path}
	return HandlerFunc(func(c *HandlerContext, in any) (any, error) {
		c.Tracer().Trace(name, args...)
		c.Workdir(path)
		return in, nil
	})
}

func Silent(s bool) Handler {
	return HandlerFunc(func(c *HandlerContext, in any) (any, error) {
		c.Silent(s)
		return in, nil
	})
}

func Tracer(t tracer.Tracer) Handler {
	return HandlerFunc(func(c *HandlerContext, in any) (any, error) {
		c.Tracer(t)
		return in, nil
	})
}

func Exec(name string, args ...string) Handler {
	return HandlerFunc(func(c *HandlerContext, in any) (any, error) {
		var r io.Reader
		if in != nil {
			var ok bool
			r, ok = in.(io.Reader)
			if !ok {
				return nil, errors.New("input is not io.Reader")
			}
		}

		c.Tracer().Trace(name, args...)

		cmd := exec.CommandContext(c.Context(), name, args...)
		cmd.Dir = c.Workdir()
		cmd.Stdin = r
		if !c.Silent() {
			cmd.Stderr = os.Stderr
		}
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			return nil, errors.Wrapf(err, "failed to get stdout pipe")
		}
		if err := cmd.Start(); err != nil {
			return nil, errors.Wrapf(err, "failed to start command")
		}
		res, err := c.Next(stdout)
		if err != nil {
			_ = cmd.Cancel()
			return nil, err
		}

		return res, cmd.Wait()
	})
}

func Tee(w ...io.Writer) Handler {
	var (
		name = "tee"
		args []string
	)
	for _, writer := range w {
		if writer == os.Stdout {
			args = append(args, "<stdout>")
			continue
		}
		if writer == os.Stderr {
			args = append(args, "<stderr>")
			continue
		}
		if f, ok := writer.(*os.File); ok {
			args = append(args, f.Name())
			continue
		}
	}
	if len(args) < len(w) {
		args = append(args, fmt.Sprintf("<%d other writers>", len(w)-len(args)))
	}
	return HandlerFunc(func(c *HandlerContext, in any) (any, error) {
		if in == nil {
			return nil, errors.New("input is nil")
		}
		var r io.Reader
		r, ok := in.(io.Reader)
		if !ok {
			return nil, errors.New("input is not io.Reader")
		}

		c.Tracer().Trace(name, args...)

		var buf bytes.Buffer
		w = append(w, &buf)
		_, err := io.Copy(io.MultiWriter(w...), r)
		return &buf, err
	})
}
