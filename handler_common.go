package yevna

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/cockroachdb/errors"

	"github.com/tlipoca9/yevna/parser"
	"github.com/tlipoca9/yevna/utils"
)

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

// ToStr returns a Handler that converts input to string.
// It tries to convert input to string using the following rules:
//   - string: returns the string.
//   - []byte: returns the string.
//   - []rune: returns the string.
//   - io.Reader: reads the reader and returns the string.
//   - any other type: uses fmt.Sprint to convert to string.
func ToStr() Handler {
	return HandlerFunc(func(c *Context, in any) (any, error) {
		switch v := in.(type) {
		case string:
			return v, nil
		case []byte:
			return string(v), nil
		case []rune:
			return string(v), nil
		case io.Reader:
			b, err := io.ReadAll(v)
			if err != nil {
				return nil, errors.Wrap(err, "failed to read from reader")
			}
			return string(b), nil
		default:
			if r, err := utils.Reader(in); err == nil {
				b, err := io.ReadAll(r)
				if err != nil {
					return nil, errors.Wrap(err, "failed to read from reader")
				}
				return string(b), nil
			}

			return fmt.Sprint(in), nil
		}
	})
}

// Chdir returns a Handler that changes the working directory.
// It uses IfExists to check if the path exists.
// It sends input to next handler.
func Chdir(path string) Handler {
	return HandlerFunc(func(c *Context, in any) (any, error) {
		if filepath.IsLocal(path) {
			path = filepath.Join(c.Workdir(), path)
		}
		_, err := os.Stat(path)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to change working directory")
		}
		c.Workdir(path)
		return in, nil
	})
}

// Silent returns a Handler that sets the silent flag.
// If silent is true, Exec will not print stderr.
// It sends original input to next handler.
func Silent(s bool) Handler {
	return HandlerFunc(func(c *Context, in any) (any, error) {
		c.Silent(s)
		return in, nil
	})
}

// Input returns a Handler that sets the input.
// It sends the input to next handler.
func Input(a any) Handler {
	return HandlerFunc(func(c *Context, _ any) (any, error) {
		return a, nil
	})
}

// Output returns a Handler that assigns the previous handler's output to a.
// If successful, it sends original input to the next handler.
func Output[T any](a *T) Handler {
	return HandlerFunc(func(c *Context, in any) (any, error) {
		switch v := in.(type) {
		case T:
			*a = v
		case *T:
			*a = *v
		default:
			return nil, errors.Newf("expected %T or %T, got %T from previous handler", *a, a, in)
		}
		return in, nil
	})
}

// Tee returns a Handler that writes to multiple writers.
// It uses io.Copy to copy the input to the writers
// and sends input to next handler.
func Tee(w ...io.Writer) Handler {
	return HandlerFunc(func(c *Context, in any) (any, error) {
		r, err := utils.Reader(in)
		if err != nil {
			return nil, err
		}

		var buf bytes.Buffer
		w = append(w, &buf)
		_, err = io.Copy(io.MultiWriter(w...), r)
		return &buf, errors.Wrap(err, "failed to copy")
	})
}

// Unmarshal returns a Handler that unmarshal the input.
// It uses the parser.Parser to unmarshal the input to v.
// It sends v to next handler.
func Unmarshal[T any](p parser.Parser, v *T) Handler {
	return HandlerFunc(func(c *Context, in any) (any, error) {
		r, err := utils.Reader(in)
		if err != nil {
			return nil, err
		}

		var b bytes.Buffer
		_, err = io.Copy(&b, r)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to copy")
		}

		err = p.Unmarshal(b.Bytes(), v)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to unmarshal")
		}

		return v, nil
	})
}

// ForEachLine returns a Handler that applies the callback function to each line.
func ForEachLine(cb func(i int, line string) string) Handler {
	return HandlerFunc(func(c *Context, in any) (any, error) {
		r, err := utils.Reader(in)
		if err != nil {
			return nil, err
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

// OpenFile returns a Handler that opens a file.
// It sends input to next handler.
func OpenFile(path string) Handler {
	return HandlerFunc(func(c *Context, in any) (any, error) {
		if filepath.IsLocal(path) {
			path = filepath.Join(c.Workdir(), path)
		}
		f, err := os.Open(path)
		if err != nil {
			return nil, errors.Wrap(err, "failed to open file")
		}

		out, err := c.Next(f)

		_ = f.Close()

		return out, err
	})
}

// WriteFile returns a Handler that writes to a file.
// If the file does not exist, it will be created.
// If the file exists, it will be truncated.
// If you want to append to a file, use AppendFile instead.
// It sends input to next handler.
func WriteFile(path ...string) Handler {
	return writeFileWithFlag(os.O_WRONLY|os.O_CREATE|os.O_TRUNC, path...)
}

// AppendFile returns a Handler that appends to a file.
// If the file does not exist, it will be created.
// If the file exists, it will be appended.
// If you want to truncate the file, use WriteFile instead.
// It sends input to next handler.
func AppendFile(path ...string) Handler {
	return writeFileWithFlag(os.O_WRONLY|os.O_CREATE|os.O_APPEND, path...)
}

// writeFileWithFlag returns a Handler that writes to a file with flag.
func writeFileWithFlag(flag int, path ...string) Handler {
	if len(path) == 0 {
		panic("no path specified")
	}
	return HandlerFunc(func(c *Context, in any) (any, error) {
		r, err := utils.Reader(in)
		if err != nil {
			return nil, err
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

		_, err = io.Copy(io.MultiWriter(ww...), r)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to write to file")
		}

		for _, f := range ff {
			_ = f.Close()
		}

		return &buf, nil
	})
}
