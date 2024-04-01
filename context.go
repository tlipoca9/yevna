package yevna

import (
	"context"
	"log/slog"
	"os"
	"os/exec"
	"sync/atomic"

	"github.com/cockroachdb/errors"

	"github.com/tlipoca9/yevna/tracer"
)

var defaultCtx atomic.Pointer[Context]

func Default() *Context {
	return defaultCtx.Load()
}

func SetDefault(c *Context) {
	defaultCtx.Store(c)
}

type Context struct {
	workDir string

	execTracer tracer.Tracer
}

type ContextOption func(*Context)

func WithWorkDir(workDir string) ContextOption {
	slog.Default()
	return func(c *Context) {
		c.workDir = workDir
	}
}

func WithExecTracer(t tracer.Tracer) ContextOption {
	return func(c *Context) {
		c.execTracer = t
	}
}

func NewContext(opts ...ContextOption) *Context {
	c := &Context{
		execTracer: tracer.NewExecTracer(os.Stderr),
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

func (c *Context) Command(ctx context.Context, name string, args ...string) *Cmd {
	cmd := exec.CommandContext(ctx, name, args...)

	// copy a new context
	cc := *c

	return &Cmd{
		ctx:     ctx,
		Context: &cc,
		cmd:     cmd,
	}
}

func (c *Context) Pipe(ctx context.Context, commands ...[]string) *Cmd {
	if len(commands) == 0 {
		panic(errors.New("no command provided"))
	}
	for i, command := range commands {
		if len(command) == 0 {
			panic(errors.Newf("No.%d command is empty", i))
		}
	}
	cmd := c.Command(ctx, commands[0][0], commands[0][1:]...)
	for i := 1; i < len(commands); i++ {
		cmd = cmd.Pipe(commands[i][0], commands[i][1:]...)
	}
	return cmd
}
