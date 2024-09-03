package yevna

import (
	"bytes"
	"io"

	"github.com/cockroachdb/errors"
	"github.com/tidwall/gjson"

	"github.com/tlipoca9/yevna/utils"
)

// Gjson returns a Handler that extracts the value using the path.
func Gjson(path string) Handler {
	return HandlerFunc(func(c *Context, in any) (any, error) {
		r, err := utils.Reader(in)
		if err != nil {
			return nil, err
		}

		b, err := io.ReadAll(r)
		if err != nil {
			return nil, errors.Wrap(err, "failed to read all")
		}

		if !gjson.ValidBytes(b) {
			return nil, errors.New("invalid json")
		}

		var buf bytes.Buffer

		value := gjson.GetBytes(b, path)
		if !value.Exists() {
			return nil, errors.New("path not found")
		}
		buf.WriteString(value.Raw)

		return &buf, nil
	})
}
