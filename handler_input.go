package yevna

// Input returns a Handler that sets the input.
// It sends the input to next handler.
func Input(a any) Handler {
	return HandlerFunc(func(c *Context, _ any) (any, error) {
		return a, nil
	})
}
