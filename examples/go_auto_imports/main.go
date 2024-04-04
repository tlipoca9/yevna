package main

import (
	"bytes"
	"context"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/cockroachdb/errors"

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

func FilterSameDirFile(files []GoFile) []GoFile {
	visit := make(map[string]GoFile)
	for _, file := range files {
		dir := filepath.Dir(file.Path)
		name := filepath.Base(file.Path)
		if name == "main.go" {
			visit[dir] = file
			continue
		}
		if _, ok := visit[dir]; !ok {
			visit[dir] = file
		}
	}

	var ret []GoFile
	for _, v := range visit {
		ret = append(ret, v)
	}

	return ret
}

func GetMissingFiles() (map[string][]string, error) {
	goFiles, err := GoMainFiles()
	if err != nil {
		return nil, err
	}
	goFiles = FilterSameDirFile(goFiles)
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
	pkgRegex := regexp.MustCompile(`^package\s+.+$`)
	importLinesRegex := regexp.MustCompile(`^import\s*\(\s*$`)
	importLineRegex := regexp.MustCompile(`^import\s+"(.+)"\s*$`)

	missingFiles, err := GetMissingFiles()
	if err != nil {
		panic(err)
	}
	for mod, files := range missingFiles {
		if len(files) == 0 {
			continue
		}
		for _, file := range files {
			var found bool
			err = yevna.Run(
				context.Background(),
				yevna.Cat(file),
				yevna.Sed(func(_ int, line string) string {
					if found {
						return line
					}
					if importLinesRegex.MatchString(line) {
						found = true
						return fmt.Sprintf("%s\n\t_ \"%s\"", line, mod)
					}
					if importLineRegex.MatchString(line) {
						found = true
						line, _ = strings.CutPrefix(line, "import")
						line = strings.TrimSpace(line)
						var buf bytes.Buffer
						buf.WriteString("import (\n")
						buf.WriteString(fmt.Sprintf("\t_ \"%s\"\n", mod))
						buf.WriteString(fmt.Sprintf("\t%s\n", line))
						buf.WriteString(")")
						return buf.String()
					}
					return line
				}),
				yevna.Sed(func(_ int, line string) string {
					if found {
						return line
					}
					if pkgRegex.MatchString(line) {
						found = true
						var buf bytes.Buffer
						buf.WriteString(line + "\n")
						buf.WriteString("import (\n")
						buf.WriteString(fmt.Sprintf("\t_ \"%s\"\n", mod))
						buf.WriteString(")")
						return buf.String()
					}
					return line
				}),
				yevna.WriteFile(file),
			)
			if err == nil && !found {
				err = errors.New("failed to add import")
			}
			if err != nil {
				panic(err)
			}
		}
	}
}
