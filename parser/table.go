package parser

import (
	"bufio"
	"errors"
	"io"
	"strings"
	"unicode"
)

type tableParser struct{}

func (t *tableParser) Parse(r io.Reader) (any, error) {
	headers := make([][2]int, 0)
	scanner := bufio.NewScanner(r)
	lines := make([]string, 0)
	maxLength := 0

	// ensure each line is the same length
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
		if len(scanner.Text()) > maxLength {
			maxLength = len(scanner.Text())
		}
	}
	if scanner.Err() != nil {
		return nil, scanner.Err()
	}
	if len(lines) == 0 {
		return nil, nil
	}
	for i, line := range lines {
		if len(line) < maxLength {
			lines[i] = line + strings.Repeat(" ", maxLength-len(line))
		}
		lines[i] += "\n"
	}

	// find the headers
	headerTxt := lines[0]
	headersMap := make([]bool, maxLength+1)
	for _, line := range lines {
		for i, c := range line {
			if !unicode.IsSpace(c) {
				headersMap[i] = true
			}
		}
	}
	i := 0
	for idx, v := range headersMap {
		if i != -1 && !v {
			headers = append(headers, [2]int{i, idx})
			i = -1
		}
		if i == -1 && v {
			i = idx
		}
	}
	if i != -1 {
		return nil, errors.New("invalid table header")
	}

	// parse the table
	ret := make([]map[string]any, 0)
	for _, line := range lines[1:] {
		v := make(map[string]any, len(headers))
		for _, h := range headers {
			v[strings.TrimSpace(headerTxt[h[0]:h[1]])] = strings.TrimSpace(line[h[0]:h[1]])
		}
		ret = append(ret, v)
	}

	return ret, nil
}

type TableParserBuilder struct {
	data tableParser
}

func (b *TableParserBuilder) Build() Parser {
	return &b.data
}

func (b *TableParserBuilder) Parse(r io.Reader) (any, error) {
	return b.Build().Parse(r)
}

func Table() *TableParserBuilder {
	return &TableParserBuilder{}
}
