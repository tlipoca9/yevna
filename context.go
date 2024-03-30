package yevna

import (
	"context"
	"os"
	"os/exec"

	"github.com/cockroachdb/errors"

	"github.com/tlipoca9/yevna/tracer"
)

type Context struct {
	ctx    context.Context
	cancel func()

	workDir string

	enableExecTrace bool
	execTracer      tracer.Tracer
}

type ContextOption func(*Context)

func WithWorkDir(workDir string) ContextOption {
	return func(c *Context) {
		c.workDir = workDir
	}
}

func WithExecTrace(enable bool) ContextOption {
	return func(c *Context) {
		c.enableExecTrace = enable
	}
}

func WithExecTracer(t tracer.Tracer) ContextOption {
	return func(c *Context) {
		c.execTracer = t
	}
}

func NewContext(ctx context.Context, opts ...ContextOption) *Context {
	cCtx, cancel := context.WithCancel(ctx)
	c := &Context{
		ctx:             cCtx,
		cancel:          cancel,
		enableExecTrace: true,
		execTracer:      tracer.NewExecTracer(os.Stderr),
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

func (c *Context) Command(name string, args ...string) *Cmd {
	cmd := exec.CommandContext(c.ctx, name, args...)

	// copy a new context
	cc := *c

	return &Cmd{
		Context: &cc,
		cmd:     cmd,
	}
}

func (c *Context) Pipe(commands ...[]string) *Cmd {
	if len(commands) == 0 {
		panic(errors.New("no command provided"))
	}
	for i, command := range commands {
		if len(command) == 0 {
			panic(errors.Newf("No.%d command is empty", i))
		}
	}
	cmd := c.Command(commands[0][0], commands[0][1:]...)
	for i := 1; i < len(commands); i++ {
		cmd = cmd.Pipe(commands[i][0], commands[i][1:]...)
	}
	return cmd
}
