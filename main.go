package main

import (
	"os"

	"github.com/urfave/cli/v2"

	"github.com/tlipoca9/yevna/cmd/helloworld"
)

func main() {
	app := cli.NewApp()

	app.Commands = append(app.Commands, helloworld.Command())

	if err := app.Run(os.Args); err != nil {
		panic(err)
	}
}
