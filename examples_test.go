package yevna_test

import (
	"context"
	"os"

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
