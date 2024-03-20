package parser

import (
	"bufio"
	"errors"
	"io"
	"strings"
	"unicode"
)

type tableParser struct{}

func Table() Parser {
	return &tableParser{}
}

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

	headerTxt := lines[0]
	i := 0
	for idx, c := range headerTxt {
		if i != -1 && unicode.IsSpace(c) {
			headers = append(headers, [2]int{i, idx})
			i = -1
		}
		if i == -1 && !unicode.IsSpace(c) {
			i = idx
		}
	}
	if i != -1 {
		return nil, errors.New("invalid table header")
	}

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
