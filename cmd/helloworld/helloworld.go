package helloworld

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-viper/mapstructure/v2"
	"github.com/tlipoca9/yevna/cmdx"
	"github.com/tlipoca9/yevna/execx"
	"github.com/tlipoca9/yevna/parser"
	"github.com/urfave/cli/v2"
	"os"
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

	err := execx.Command(ctx, "echo", "Hello World").Run()
	if err != nil {
		return err
	}

	ezaResult := make([]map[string]any, 0)
	err = execx.Command(ctx, "eza", "-l", "--header").
		WithParser(parser.Table()).
		WithDecoderConfig(mapstructure.DecoderConfig{Result: &ezaResult}).
		Run()
	if err != nil {
		return err
	}
	buf, _ := json.MarshalIndent(ezaResult, "", "  ")
	os.Stdout.Write(buf)
	fmt.Println()

	err = execx.Command(ctx, "go", "mod", "tidy").Run()
	if err != nil {
		return err
	}

	isRepoInit, err := cmdx.FileExists(".git")
	if err != nil {
		return err
	}
	if !isRepoInit {
		err = execx.Command(ctx, "git", "init").Run()
		if err != nil {
			return err
		}
	}

	type gitConfig struct {
		InitBranch string `json:"init.defaultbranch"`
	}
	var conf gitConfig
	err = execx.Command(ctx, "git", "config", "--global", "--list").
		WithParser(parser.Dotenv()).
		WithDecoderConfig(mapstructure.DecoderConfig{TagName: "json", Result: &conf}).
		Run()
	if err != nil {
		return err
	}
	if conf.InitBranch != "main" {
		err = execx.Command(ctx, "git", "config", "--global", "init.defaultBranch", "main").Run()
		if err != nil {
			return err
		}
	}

	err = execx.Command(ctx, "git", "add", ".").Run()
	if err != nil {
		return err
	}

	return execx.Command(ctx, "git", "commit", "-m", "init").Run()
}
