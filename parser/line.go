package parser

import (
	"bufio"
	"io"
)

type lineParser struct {
	MatchFunc func(line string) bool
}

func (p *lineParser) Parse(r io.Reader) (any, error) {
	fn := p.MatchFunc
	if fn == nil {
		fn = func(_ string) bool { return true }
	}

	scanner := bufio.NewScanner(r)
	var lines []string
	for scanner.Scan() {
		line := scanner.Text()
		if fn(line) {
			lines = append(lines, line)
		}
	}

	return lines, scanner.Err()
}

type LineParserBuilder struct {
	data lineParser
}

func (b *LineParserBuilder) MatchFunc(f func(line string) bool) *LineParserBuilder {
	b.data.MatchFunc = f
	return b
}

func (b *LineParserBuilder) Build() Parser {
	return &b.data
}

func (b *LineParserBuilder) Parse(r io.Reader) (any, error) {
	return b.Build().Parse(r)
}

func Line() *LineParserBuilder {
	return &LineParserBuilder{}
}
