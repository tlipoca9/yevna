package main

import (
	"context"
	"fmt"

	"github.com/tlipoca9/yevna"
	"github.com/tlipoca9/yevna/parser"
)

func GoFiles() ([]string, error) {
	var ret []string
	err := yevna.Run(
		context.Background(),
		yevna.Exec("go", "list",
			"-f",
			`{{ range .GoFiles }}{{ printf "%s/%s\n" $.Dir . }}{{ end }}`,
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
	files, err := GoFiles()
	if err != nil {
		panic(err)
	}
	for _, file := range files {
		fmt.Println(file)
	}
}
