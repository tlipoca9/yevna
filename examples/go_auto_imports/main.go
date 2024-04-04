package main

import (
	"context"
	"fmt"
	"regexp"

	"github.com/tlipoca9/yevna"
	"github.com/tlipoca9/yevna/parser"
)

var autoModules = []string{
	"go.uber.org/automaxprocs",
	"github.com/KimMachineGun/automemlimit",
}

type GoFile struct {
	Imports []string `json:"imports"`
	Path    string   `json:"path"`
}

func GoMainFiles() ([]GoFile, error) {
	var ret []GoFile
	err := yevna.Run(
		context.Background(),
		yevna.Exec(
			"go",
			"list",
			"-f",
			`{{ range .GoFiles }}{{ if eq $.Name "main" }}{{ printf "- imports: %s\n  path: \"%s/%s\"\n" $.Imports $.Dir . }}{{ end }}{{ end }}`,
			"./...",
		),
		yevna.Unmarshal(parser.YAML(), &ret),
	)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func GetMissingFiles() (map[string][]string, error) {
	goFiles, err := GoMainFiles()
	if err != nil {
		return nil, err
	}
	missingFiles := make(map[string][]string, len(autoModules))
	for _, mod := range autoModules {
		missingFiles[mod] = []string{}
		for _, file := range goFiles {
			found := false
			for _, imp := range file.Imports {
				if imp == mod {
					found = true
					break
				}
			}
			if !found {
				missingFiles[mod] = append(missingFiles[mod], file.Path)
			}
		}
	}

	return missingFiles, nil
}

func main() {
	importLineRegex := regexp.MustCompile(`^import\s*\(\s*$`)

	missingFiles, err := GetMissingFiles()
	if err != nil {
		panic(err)
	}
	for mod, files := range missingFiles {
		if len(files) == 0 {
			continue
		}
		for _, file := range files {
			err = yevna.Run(
				context.Background(),
				yevna.Cat(file),
				yevna.Sed(func(line string) string {
					if importLineRegex.MatchString(line) {
						return fmt.Sprintf("%s\n\t_ \"%s\"", line, mod)
					}
					return line
				}),
				yevna.WriteFile(file),
			)
			if err != nil {
				panic(err)
			}
		}
	}
}
