package parser

import (
	"bufio"
	"io"
)

type SepParserOptions struct {
	SplitFunc bufio.SplitFunc
	Filter    func(token string) bool
}

type SepParser struct {
	opt SepParserOptions
}

func (p *SepParser) SplitFunc(f bufio.SplitFunc) *SepParser {
	p.opt.SplitFunc = f
	return p
}

func (p *SepParser) Filter(f func(token string) bool) *SepParser {
	p.opt.Filter = f
	return p
}

func (p *SepParser) lazyInit() {
	if p.opt.SplitFunc == nil {
		p.opt.SplitFunc = bufio.ScanWords
	}
	if p.opt.Filter == nil {
		p.opt.Filter = func(_ string) bool { return true }
	}
}

func (p *SepParser) Parse(r io.Reader) (any, error) {
	p.lazyInit()

	scanner := bufio.NewScanner(r)
	scanner.Split(p.opt.SplitFunc)
	var tokens []string
	for scanner.Scan() {
		token := scanner.Text()
		if p.opt.Filter(token) {
			tokens = append(tokens, token)
		}
	}

	return tokens, scanner.Err()
}

// Sep returns a new SepParser
func Sep() *SepParser {
	return &SepParser{}
}

// Line is the shortcut for Sep().SplitFunc(bufio.ScanLines)
func Line() *SepParser {
	return Sep().SplitFunc(bufio.ScanLines)
}
