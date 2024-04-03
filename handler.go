package yevna

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/cockroachdb/errors"

	"github.com/tlipoca9/yevna/parser"
	"github.com/tlipoca9/yevna/tracer"
)

// HandlersChain defines a Handler slice
type HandlersChain []Handler

type Handler interface {
	Handle(c *Context, in any) (any, error)
}

// HandlerFunc defines a function type that implements Handler
type HandlerFunc func(c *Context, in any) (any, error)

func (f HandlerFunc) Handle(c *Context, in any) (any, error) {
	return f(c, in)
}

// Cd returns a Handler that changes the working directory
func Cd(path string) Handler {
	name := "cd"
	args := []string{path}
	return HandlerFunc(func(c *Context, in any) (any, error) {
		c.Tracer().Trace(name, args...)
		c.Workdir(path)
		return in, nil
	})
}

// Silent returns a Handler that sets the silent flag
func Silent(s bool) Handler {
	return HandlerFunc(func(c *Context, in any) (any, error) {
		c.Silent(s)
		return in, nil
	})
}

// Tracer returns a Handler that sets the tracer
func Tracer(t tracer.Tracer) Handler {
	return HandlerFunc(func(c *Context, in any) (any, error) {
		c.Tracer(t)
		return in, nil
	})
}

// Echo returns a Handler that echoes the input
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

	return HandlerFunc(func(c *Context, _ any) (any, error) {
		c.Tracer().Trace(name, args...)
		return r, nil
	})
}

// Exec returns a Handler that executes a command
func Exec(name string, args ...string) Handler {
	return HandlerFunc(func(c *Context, in any) (any, error) {
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

// Tee returns a Handler that writes to multiple writers
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
	return HandlerFunc(func(c *Context, in any) (any, error) {
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

// Unmarshal returns a Handler that unmarshal the input
func Unmarshal(p parser.Parser, v any) Handler {
	var (
		name = "unmarshal"
		args = []string{fmt.Sprintf("%T", v)}
	)
	return HandlerFunc(func(c *Context, in any) (any, error) {
		if in == nil {
			return nil, errors.New("input is nil")
		}
		var b bytes.Buffer
		switch r := in.(type) {
		case io.Reader:
			_, err := io.Copy(&b, r)
			if err != nil {
				return nil, errors.Wrapf(err, "failed to copy")
			}
		case []byte:
			_, err := b.Write(r)
			if err != nil {
				return nil, errors.Wrapf(err, "failed to write")
			}
		case string:
			_, err := b.WriteString(r)
			if err != nil {
				return nil, errors.Wrapf(err, "failed to write")
			}
		default:
			return nil, errors.New("input is not io.Reader, []byte or string")
		}

		c.Tracer().Trace(name, args...)

		err := p.Unmarshal(b.Bytes(), v)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to unmarshal")
		}

		return v, nil
	})
}
