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

// ErrorHandler returns a Handler that handles error.
// It prints the error using fmt.Printf("%+v\n", err).
func ErrorHandler() Handler {
	return HandlerFunc(func(c *Context, in any) (any, error) {
		out, err := c.Next(in)
		if err != nil {
			fmt.Printf("%+v\n", err)
		}
		return out, err
	})
}

// Recover returns a Handler that recovers from panic.
// It wraps the panic error using errors.Wrap.
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

// Cd returns a Handler that changes the working directory.
// It uses IfExists to check if the path exists.
// It sends input to next handler.
func Cd(path string) Handler {
	name := "~cd"
	args := []string{path}
	return HandlerFunc(func(c *Context, in any) (any, error) {
		c.Tracer().Trace(name, args...)
		_, err := IfExists(path).Handle(c, nil)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to change working directory")
		}
		c.Workdir(path)
		return in, nil
	})
}

// Silent returns a Handler that sets the silent flag.
// If silent is true, Exec will not print stderr.
// If you want to disable tracing, use Tracer(tracer.Discard) instead.
// It sends input to next handler.
func Silent(s bool) Handler {
	return HandlerFunc(func(c *Context, in any) (any, error) {
		c.Silent(s)
		return in, nil
	})
}

// Tracer returns a Handler that sets the tracer.
// If you want to disable tracing, use Tracer(tracer.Discard).
// It sends input to next handler.
func Tracer(t tracer.Tracer) Handler {
	return HandlerFunc(func(c *Context, in any) (any, error) {
		c.Tracer(t)
		return in, nil
	})
}

// WithReader returns a Handler that sets the reader.
// If you want to read from a file, use Cat instead.
// If you want to read from a string, use Echo instead.
// It drops the input and sends io.Reader to next handler.
func WithReader(r io.Reader) Handler {
	var (
		name = "~with_reader"
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

// Echo returns a Handler that echoes the string.
// If you want to read from a file, use Cat instead.
// If you want to read from a reader, use WithReader instead.
// It drops the input and sends string(converted to io.Reader) to next handler.
func Echo(s string) Handler {
	return HandlerFunc(func(c *Context, _ any) (any, error) {
		c.Tracer().Trace("~echo", s)
		return bytes.NewBufferString(s), nil
	})
}

// Exec returns a Handler that executes a command.
// It uses exec.CommandContext to execute the command.
//   - stdin is set to the input.
//   - stdout is sent to next handler.
//   - stderr is sent to os.Stderr if silent is false.
//
// It starts the command and waits after the next handler is called.
func Exec(name string, args ...string) Handler {
	return HandlerFunc(func(c *Context, in any) (any, error) {
		c.Tracer().Trace(name, args...)

		var r io.Reader
		if in != nil {
			var ok bool
			r, ok = in.(io.Reader)
			if !ok {
				return nil, errors.New("input is not io.Reader")
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

// Tee returns a Handler that writes to multiple writers.
// It uses io.Copy to copy the input to the writers
// and sends input to next handler.
func Tee(w ...io.Writer) Handler {
	var (
		name = "~tee"
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
		c.Tracer().Trace(name, args...)

		if in == nil {
			return nil, errors.New("input is nil")
		}
		var r io.Reader
		r, ok := in.(io.Reader)
		if !ok {
			return nil, errors.New("input is not io.Reader")
		}

		var buf bytes.Buffer
		w = append(w, &buf)
		_, err := io.Copy(io.MultiWriter(w...), r)
		return &buf, errors.Wrap(err, "failed to copy")
	})
}

// writeFileWithFlag returns a Handler that writes to a file with flag.
func writeFileWithFlag(name string, flag int, path ...string) Handler {
	if len(path) == 0 {
		panic("no path specified")
	}
	return HandlerFunc(func(c *Context, in any) (any, error) {
		c.Tracer().Trace(name, path...)

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

		// copy to buffer
		var buf bytes.Buffer
		ww = append(ww, &buf)

		_, err := io.Copy(io.MultiWriter(ww...), r)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to write to file")
		}

		for _, f := range ff {
			_ = f.Close()
		}

		return &buf, nil
	})
}

// WriteFile returns a Handler that writes to a file.
// If the file does not exist, it will be created.
// If the file exists, it will be truncated.
// If you want to append to a file, use AppendFile instead.
// It sends input to next handler.
func WriteFile(path ...string) Handler {
	return writeFileWithFlag("~write_file", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, path...)
}

// AppendFile returns a Handler that appends to a file.
// If the file does not exist, it will be created.
// If the file exists, it will be appended.
// If you want to truncate the file, use WriteFile instead.
// It sends input to next handler.
func AppendFile(path ...string) Handler {
	return writeFileWithFlag("~append_file", os.O_WRONLY|os.O_CREATE|os.O_APPEND, path...)
}

// Unmarshal returns a Handler that unmarshal the input.
// It uses the parser.Parser to unmarshal the input to v.
// It sends v to next handler.
func Unmarshal(p parser.Parser, v any) Handler {
	var (
		name = "~unmarshal"
		args = []string{fmt.Sprintf("<%T>", v)}
	)
	return HandlerFunc(func(c *Context, in any) (any, error) {
		c.Tracer().Trace(name, args...)

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

		err := p.Unmarshal(b.Bytes(), v)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to unmarshal")
		}

		return v, nil
	})
}

// IfExists returns a Handler that checks if the file exists.
// If the file does not exist, it returns an error.
// If the file exists, it sends input to next handler.
func IfExists(path ...string) Handler {
	name := "~if_exists"
	args := path
	return HandlerFunc(func(c *Context, in any) (any, error) {
		c.Tracer().Trace(name, args...)

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

// Cat returns a Handler that reads a file.
// It drops the input and sends file content (io.Reader) to next handler.
func Cat(path string) Handler {
	name := "~cat"
	args := []string{path}
	return HandlerFunc(func(c *Context, _ any) (any, error) {
		c.Tracer().Trace(name, args...)

		if filepath.IsLocal(path) {
			path = filepath.Join(c.Workdir(), path)
		}
		f, err := os.Open(path)
		if err != nil {
			return nil, errors.Wrap(err, "failed to open file")
		}
		return f, nil
	})
}

// Sed returns a Handler that applies the callback function to each line.
func Sed(cb func(i int, line string) string) Handler {
	name := "~sed"
	args := []string{fmt.Sprintf("<%T>", cb)}
	return HandlerFunc(func(c *Context, in any) (any, error) {
		c.Tracer().Trace(name, args...)

		if in == nil {
			return nil, errors.New("input is nil")
		}
		var r io.Reader
		r, ok := in.(io.Reader)
		if !ok {
			return nil, errors.New("input is not io.Reader")
		}

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
