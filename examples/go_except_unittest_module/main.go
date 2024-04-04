package main

import (
	"context"
	"fmt"

	"github.com/tlipoca9/yevna"
	"github.com/tlipoca9/yevna/parser"
)

func GoExceptUnittestModule() ([]string, error) {
	var ret []string
	err := yevna.Run(
		context.Background(),
		yevna.Exec(
			"go",
			"list",
			"-f", `{{ if not (or .XTestGoFiles .TestGoFiles) }}{{ .Dir }}{{ end }}`,
			"./...",
		),
		yevna.Unmarshal(parser.Line(), &ret),
	)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func main() {
	dirs, err := GoExceptUnittestModule()
	if err != nil {
		panic(err)
	}
	for _, d := range dirs {
		fmt.Println(d)
	}
}
