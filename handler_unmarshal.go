package yevna

import (
	"bytes"
	"io"

	"github.com/cockroachdb/errors"

	"github.com/tlipoca9/yevna/parser"
	"github.com/tlipoca9/yevna/utils"
)

// Unmarshal returns a Handler that unmarshal the input.
// It uses the parser.Parser to unmarshal the input to v.
// It sends v to next handler.
func Unmarshal(p parser.Parser, v any) Handler {
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
