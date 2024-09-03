package yevna

import "fmt"

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
