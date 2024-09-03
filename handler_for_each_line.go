package yevna

import (
	"bufio"
	"bytes"

	"github.com/cockroachdb/errors"

	"github.com/tlipoca9/yevna/utils"
)

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
