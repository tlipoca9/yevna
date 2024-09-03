package yevna

import (
	"bytes"
	"io"

	"github.com/cockroachdb/errors"

	"github.com/tlipoca9/yevna/utils"
)

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
