package parser

import (
	"bufio"
	"errors"
	"io"
	"strings"
	"unicode"
)

type tableParser struct {
	sep     func(rune) bool
	skip    func(i int, line string) bool
	headers [][2]int
	trimKey func(rune) bool
	trimVal func(rune) bool
}

// Parse is a method that implements the Parser interface
func (t *tableParser) Parse(r io.Reader) (any, error) {
	sep, maxLength := " ", 0

	// ensure each line is the same length

	// get the maximum length of the line
	scanner := bufio.NewScanner(r)
	lines := make([]string, 0)
	for i := 0; scanner.Scan(); i++ {
		if t.skip(i, scanner.Text()) {
			continue
		}
		lines = append(lines, scanner.Text())
		if len(scanner.Text()) > maxLength {
			maxLength = len(scanner.Text())
		}
	}
	if scanner.Err() != nil {
		return nil, scanner.Err()
	}
	if len(lines) == 0 {
		return []map[string]any{}, nil
	}

	headerTxt := lines[0]

	// get the separator
	for _, c := range headerTxt {
		if t.sep(c) {
			sep = string(c)
			break
		}
	}

	// ensure each line is the same length
	// use the separator to fill the line
	for i, line := range lines {
		if len(line) < maxLength {
			lines[i] = line + strings.Repeat(sep, maxLength-len(line))
		}
	}

	// find the headers
	headersMap := make([]bool, maxLength+1)
	for _, line := range lines {
		for i, c := range line {
			if !t.sep(c) {
				headersMap[i] = true
			}
		}
	}
	i, headers := 0, make([][2]int, 0)
	for idx, v := range headersMap {
		if i != -1 && !v {
			i, headers = -1, append(headers, [2]int{i, idx})
		} else if i == -1 && v {
			i = idx
		}
	}
	if i != -1 {
		return nil, errors.New("invalid table header")
	}

	// parse the table
	ret := make([]map[string]any, 0, len(lines)-1)
	for _, line := range lines[1:] {
		obj := make(map[string]any, len(headers))
		for _, h := range headers {
			k := headerTxt[h[0]:min(h[1], len(headerTxt))]
			v := line[h[0]:min(h[1], len(line))]
			obj[strings.TrimFunc(k, t.trimKey)] = strings.TrimFunc(v, t.trimVal)
		}
		ret = append(ret, obj)
	}

	return ret, nil
}

// TableParserBuilder is a builder for the TableParser
type TableParserBuilder struct {
	headerTxt string

	data tableParser
}

// Sep sets the separator for the table
func (b *TableParserBuilder) Sep(sep func(rune) bool) *TableParserBuilder {
	b.data.sep = sep
	return b
}

// Skip sets the skip function for the table
func (b *TableParserBuilder) Skip(skip func(i int, line string) bool) *TableParserBuilder {
	b.data.skip = skip
	return b
}

// Header sets the header for the table
func (b *TableParserBuilder) Header(header string) *TableParserBuilder {
	b.headerTxt = header
	return b
}

// TrimKey sets the trim key function for the table
func (b *TableParserBuilder) TrimKey(trim func(rune) bool) *TableParserBuilder {
	b.data.trimKey = trim
	return b
}

// TrimVal sets the trim value function for the table
func (b *TableParserBuilder) TrimVal(trim func(rune) bool) *TableParserBuilder {
	b.data.trimVal = trim
	return b
}

// Build returns a new Parser
func (b *TableParserBuilder) Build() Parser {
	if b.data.sep == nil {
		b.data.sep = unicode.IsSpace
	}
	if b.data.skip == nil {
		b.data.skip = func(_ int, _ string) bool { return false }
	}
	if b.headerTxt != "" {
		b.data.headers = make([][2]int, 0)
		headersMap := make([]bool, len(b.headerTxt)+1)
		for i, c := range b.headerTxt {
			if !b.data.sep(c) {
				headersMap[i] = true
			}
		}
		i := 0
		for idx, v := range headersMap {
			if i != -1 && !v {
				b.data.headers = append(b.data.headers, [2]int{i, idx})
				i = -1
			}
			if i == -1 && v {
				i = idx
			}
		}
		if i != -1 {
			panic("invalid table header")
		}
	}
	if b.data.trimKey == nil {
		b.data.trimKey = unicode.IsSpace
	}
	if b.data.trimVal == nil {
		b.data.trimVal = unicode.IsSpace
	}
	return &b.data
}

// Parse is a magic method that implements the Parser interface
func (b *TableParserBuilder) Parse(r io.Reader) (any, error) {
	return b.Build().Parse(r)
}

// Table returns a new TableParserBuilder to parse a table
// It parses a table from an io.Reader
// The table must be in the following format:
//
// ```
// header1 header2 header3
// value1  value2  value3
// value4  value5  value6
// ```
//
// The headers and the values must be separated by unicode space
func Table() *TableParserBuilder {
	return &TableParserBuilder{}
}
