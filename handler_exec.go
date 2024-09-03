package yevna

import (
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/cockroachdb/errors"
	"mvdan.cc/sh/v3/shell"

	"github.com/tlipoca9/yevna/utils"
)

// Exec returns a Handler that executes a command.
// It uses exec.CommandContext to execute the command.
//   - stdin is set to the input.
//   - stdout is sent to next handler.
//   - stderr is sent to os.Stderr if silent is false.
//
// It starts the command and waits after the next handler is called.
func Exec(name string, args ...string) Handler {
	return HandlerFunc(func(c *Context, in any) (any, error) {
		var (
			r   io.Reader
			err error
		)
		if in != nil {
			r, err = utils.Reader(in)
			if err != nil {
				return nil, err
			}
		}

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
		if err = cmd.Start(); err != nil {
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

// Execs returns a Handler that executes a command.
// It uses shell.Fields to parse the command.
// It is a shortcut for Exec(shell.Fields(cmd)).
func Execs(cmd string) Handler {
	return HandlerFunc(func(c *Context, in any) (any, error) {
		args, err := shell.Fields(cmd, nil)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to parse command")
		}

		return Exec(args[0], args[1:]...).Handle(c, in)
	})
}

// Execf returns a Handler that executes a command.
// It is a shortcut for Execs(fmt.Sprintf(format, a...)).
func Execf(format string, a ...any) Handler {
	return Execs(fmt.Sprintf(format, a...))
}
