package yevna

import "github.com/cockroachdb/errors"

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
