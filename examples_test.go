package yevna_test

import (
	"context"
	"os"

	"github.com/cockroachdb/errors"

	"github.com/tlipoca9/yevna"
)

func ExampleRun() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := yevna.Run(
		ctx,
		yevna.Exec("echo", "Hello, World!"),
		yevna.Tee(os.Stdout),
	)
	if err != nil {
		panic(err)
	}
	// Output:
	// Hello, World!
}

func ExamplePanic() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Default context is configured with Recover and ErrorHandler
	err := yevna.Run(
		ctx,
		yevna.HandlerFunc(func(_ *yevna.Context, _ any) (any, error) {
			panic(errors.New("something went wrong"))
		}),
	)
	if err == nil {
		panic("error expected")
	}
}
