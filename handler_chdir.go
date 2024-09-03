package yevna

import (
	"os"
	"path/filepath"

	"github.com/cockroachdb/errors"
)

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
