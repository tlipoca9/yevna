package yevna

import (
	"bytes"
	"io"
	"os"
	"path/filepath"

	"github.com/cockroachdb/errors"

	"github.com/tlipoca9/yevna/utils"
)

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
