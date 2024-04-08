package yevna

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"slices"

	"github.com/cockroachdb/errors"
	"mvdan.cc/sh/v3/shell"

	"github.com/tlipoca9/yevna/parser"
	"github.com/tlipoca9/yevna/tracer"
)

// HandlersChain defines a Handler slice
type HandlersChain []Handler

func (c HandlersChain) Copy() HandlersChain {
	return slices.Clone(c)
}

type Handler interface {
	Handle(c *Context, in any) (any, error)
}

// HandlerFunc defines a function type that implements Handler
type HandlerFunc func(c *Context, in any) (any, error)

func (f HandlerFunc) Handle(c *Context, in any) (any, error) {
	return f(c, in)
}

func ErrorHandler() Handler {
	return HandlerFunc(func(c *Context, in any) (any, error) {
		out, err := c.Next(in)
		if err != nil {
			fmt.Printf("%+v\n", err)
		}
		return out, err
	})
}

// Recover returns a Handler that recovers from panic
func Recover() Handler {
	return HandlerFunc(func(c *Context, in any) (_ any, err error) {
		defer func() {
			if r := recover(); r != nil {
				if e, ok := r.(error); ok {
					err = errors.Wrap(e, "recovered from panic")
				} else {
					err = errors.Newf("recovered from panic: %v", r)
				}
			}
		}()
		return c.Next(in)
	})
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

// WithReader returns a Handler that sets the reader
func WithReader(r io.Reader) Handler {
	var (
		name = "with_reader"
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

// Echo returns a Handler that echoes the string
func Echo(s string) Handler {
	return HandlerFunc(func(c *Context, _ any) (any, error) {
		c.Tracer().Trace("echo", s)
		return bytes.NewBufferString(s), nil
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

// Execs returns a Handler that executes a command
func Execs(cmd string) Handler {
	return HandlerFunc(func(c *Context, in any) (any, error) {
		args, err := shell.Fields(cmd, nil)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to parse command")
		}

		return Exec(args[0], args[1:]...).Handle(c, in)
	})
}

// Execf returns a Handler that executes a command
// It is a shortcut for Execs(fmt.Sprintf(format, a...))
func Execf(format string, a ...any) Handler {
	return Execs(fmt.Sprintf(format, a...))
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
		return &buf, errors.Wrap(err, "failed to copy")
	})
}

// writeFileWithFlag returns a Handler that writes to a file with flag
func writeFileWithFlag(name string, flag int, path ...string) Handler {
	if len(path) == 0 {
		panic("no path specified")
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

		ff := make([]*os.File, 0, len(path))
		ww := make([]io.Writer, 0, len(path))
		for i := range path {
			if filepath.IsLocal(path[i]) {
				path[i] = filepath.Join(c.Workdir(), path[i])
			}
			f, err := os.OpenFile(path[i], flag, 0644)
			if err != nil {
				return nil, errors.Wrapf(err, "failed to open file")
			}
			ff = append(ff, f)
			ww = append(ww, f)
		}

		c.Tracer().Trace(name, path...)

		_, err := io.Copy(io.MultiWriter(ww...), r)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to write to file")
		}

		for _, f := range ff {
			_ = f.Close()
		}

		return nil, nil
	})
}

// WriteFile returns a Handler that writes to a file
func WriteFile(path ...string) Handler {
	return writeFileWithFlag("write_file", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, path...)
}

// AppendFile returns a Handler that appends to a file
func AppendFile(path ...string) Handler {
	return writeFileWithFlag("append_file", os.O_WRONLY|os.O_CREATE|os.O_APPEND, path...)
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

// IfExists returns a Handler that checks if the file exists
func IfExists(path ...string) Handler {
	return HandlerFunc(func(c *Context, in any) (any, error) {
		for _, p := range path {
			if filepath.IsLocal(p) {
				p = filepath.Join(c.Workdir(), p)
			}
			_, err := os.Stat(p)
			if err != nil {
				return nil, errors.Wrapf(err, "file %s does not exist", p)
			}
		}
		return in, nil
	})
}

// Cat returns a Handler that reads a file
func Cat(path string) Handler {
	return HandlerFunc(func(c *Context, _ any) (any, error) {
		if filepath.IsLocal(path) {
			path = filepath.Join(c.Workdir(), path)
		}
		c.Tracer().Trace("cat", path)
		f, err := os.Open(path)
		if err != nil {
			return nil, errors.Wrap(err, "failed to open file")
		}
		return f, nil
	})
}

// Sed returns a Handler that applies the callback function to each line
func Sed(cb func(i int, line string) string) Handler {
	name := "sed"
	args := []string{fmt.Sprintf("<%T>", cb)}
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
		scanner := bufio.NewScanner(r)
		scanner.Split(bufio.ScanLines)
		for i := 0; scanner.Scan(); i++ {
			line := cb(i, scanner.Text())
			buf.WriteString(line)
			buf.WriteByte('\n')
		}
		if err := scanner.Err(); err != nil {
			return nil, errors.Wrap(err, "failed to scan")
		}

		return &buf, nil
	})
}
