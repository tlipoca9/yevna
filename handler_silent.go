package yevna

// Silent returns a Handler that sets the silent flag.
// If silent is true, Exec will not print stderr.
// It sends input to next handler.
func Silent(s bool) Handler {
	return HandlerFunc(func(c *Context, in any) (any, error) {
		c.Silent(s)
		return in, nil
	})
}
