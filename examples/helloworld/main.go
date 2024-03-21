package main

import (
	"context"
	"github.com/tlipoca9/yevna"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := yevna.Command(ctx, "echo", "Hello World").Run().Err
	if err != nil {
		panic(err)
	}
}
