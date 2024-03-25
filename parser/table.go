package parser

import (
	"bufio"
	"io"
	"strings"
	"unicode"

	"github.com/cockroachdb/errors"
)

type TableHeader string

func (h TableHeader) Index(isSeperator func(rune) bool, lines []string) ([][2]int, error) {
	ret := make([][2]int, 0)
	headersMap := make([]bool, len(h)+1)
	for i, c := range h {
		if !isSeperator(c) {
			headersMap[i] = true
		}
	}
	for _, line := range lines {
		if len(headersMap) < len(line)+1 {
			headersMap = append(headersMap, make([]bool, len(line)+1-len(headersMap))...)
		}
		for i, c := range line {
			if !isSeperator(c) {
				headersMap[i] = true
			}
		}
	}

	i := 0
	for idx, v := range headersMap {
		if i != -1 && !v {
			ret = append(ret, [2]int{i, idx})
			i = -1
		}
		if i == -1 && v {
			i = idx
		}
	}
	if i != -1 {
		return nil, errors.New("invalid table header")
	}
	return ret, nil
}

type TableParserOptions struct {
	SeperatorFunc    func(rune) bool
	Filter           func(i int, line string) bool
	KeyTransformFunc func(string) string
	ValTransformFunc func(string) string
	HeaderTxt        string
}

// TableParser is a builder for the TableParser
type TableParser struct {
	opt TableParserOptions
}

func (b *TableParser) WithSeperatorFunc(f func(rune) bool) *TableParser {
	b.opt.SeperatorFunc = f
	return b
}

func (b *TableParser) WithFilter(f func(i int, line string) bool) *TableParser {
	b.opt.Filter = f
	return b
}

func (b *TableParser) WithKeyTransformFunc(f func(string) string) *TableParser {
	b.opt.KeyTransformFunc = f
	return b
}

func (b *TableParser) WithValTransformFunc(f func(string) string) *TableParser {
	b.opt.ValTransformFunc = f
	return b
}

func (b *TableParser) WithHeader(header string) *TableParser {
	b.opt.HeaderTxt = header
	return b
}

func (b *TableParser) lazyInit() {
	if b.opt.SeperatorFunc == nil {
		b.opt.SeperatorFunc = unicode.IsSpace
	}
	if b.opt.Filter == nil {
		b.opt.Filter = func(_ int, _ string) bool { return true }
	}
	if b.opt.KeyTransformFunc == nil {
		b.opt.KeyTransformFunc = func(s string) string { return strings.TrimFunc(s, b.opt.SeperatorFunc) }
	}
	if b.opt.ValTransformFunc == nil {
		b.opt.ValTransformFunc = func(s string) string { return strings.TrimFunc(s, b.opt.SeperatorFunc) }
	}
}

// Parse is a magic method that implements the Parser interface
func (b *TableParser) Parse(r io.Reader) (any, error) {
	b.lazyInit()

	// get the maximum length of the line
	scanner := bufio.NewScanner(r)
	lines := make([]string, 0)
	for i := 0; scanner.Scan(); i++ {
		if b.opt.Filter(i, scanner.Text()) {
			lines = append(lines, scanner.Text())
		}
	}
	if scanner.Err() != nil {
		return nil, scanner.Err()
	}
	if len(lines) == 0 {
		return []map[string]any{}, nil
	}

	headerTxt := b.opt.HeaderTxt
	if len(headerTxt) == 0 {
		headerTxt, lines = lines[0], lines[1:]
		if len(lines) == 0 {
			return []map[string]any{}, nil
		}
	}
	header := TableHeader(headerTxt)
	headerIndex, err := header.Index(b.opt.SeperatorFunc, lines)
	if err != nil {
		return nil, err
	}

	// parse the table
	ret := make([]map[string]any, 0, len(lines))
	for _, line := range lines {
		obj := make(map[string]any, len(headerIndex))
		for _, h := range headerIndex {
			k := b.opt.KeyTransformFunc(headerTxt[h[0]:min(h[1], len(headerTxt))])
			v := b.opt.ValTransformFunc(line[h[0]:min(h[1], len(line))])
			obj[k] = v
		}
		ret = append(ret, obj)
	}

	return ret, nil
}

// Table returns a new TableParser to parse a table
// It parses a table from an io.Reader
func Table() *TableParser {
	return &TableParser{}
}
