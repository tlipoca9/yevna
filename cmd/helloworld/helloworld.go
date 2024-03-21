package helloworld

import (
	"context"

	"github.com/urfave/cli/v2"

	"github.com/tlipoca9/yevna/execx"
)

func Command() *cli.Command {
	return &cli.Command{
		Name:   "helloworld",
		Action: Action,
	}
}

func Action(cCtx *cli.Context) error {
	ctx, cancel := context.WithCancel(cCtx.Context)
	defer cancel()

	return execx.Command(ctx, "echo", "Hello World").Run().Err
}
