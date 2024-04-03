package parser

import (
	"bufio"
	"bytes"
	"strings"
	"unicode"

	"github.com/cockroachdb/errors"
	"github.com/go-viper/mapstructure/v2"
)

type TableHeader string

func (h TableHeader) Index(isSeparator func(rune) bool, lines []string) ([][2]int, error) {
	ret := make([][2]int, 0)
	headersMap := make([]bool, len(h)+1)
	for i, c := range h {
		if !isSeparator(c) {
			headersMap[i] = true
		}
	}
	for _, line := range lines {
		if len(headersMap) < len(line)+1 {
			headersMap = append(headersMap, make([]bool, len(line)+1-len(headersMap))...)
		}
		for i, c := range line {
			if !isSeparator(c) {
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

// TableParser is a builder for the TableParser
type TableParser struct {
	conf             *mapstructure.DecoderConfig
	separatorFunc    func(rune) bool
	filter           func(i int, line string) bool
	keyTransformFunc func(string) string
	valTransformFunc func(string) string
	headerTxt        string
}

func (p *TableParser) WithSeparatorFunc(f func(rune) bool) *TableParser {
	p.separatorFunc = f
	return p
}

func (p *TableParser) WithFilter(f func(i int, line string) bool) *TableParser {
	p.filter = f
	return p
}

func (p *TableParser) WithKeyTransformFunc(f func(string) string) *TableParser {
	p.keyTransformFunc = f
	return p
}

func (p *TableParser) WithValTransformFunc(f func(string) string) *TableParser {
	p.valTransformFunc = f
	return p
}

func (p *TableParser) WithHeader(header string) *TableParser {
	p.headerTxt = header
	return p
}

func (p *TableParser) lazyInit() {
	if p.separatorFunc == nil {
		p.separatorFunc = unicode.IsSpace
	}
	if p.filter == nil {
		p.filter = func(_ int, _ string) bool { return true }
	}
	if p.keyTransformFunc == nil {
		p.keyTransformFunc = func(s string) string { return strings.TrimFunc(s, p.separatorFunc) }
	}
	if p.valTransformFunc == nil {
		p.valTransformFunc = func(s string) string { return strings.TrimFunc(s, p.separatorFunc) }
	}
	if p.conf == nil {
		p.conf = &mapstructure.DecoderConfig{TagName: "json"}
	}
	if p.conf.TagName == "" {
		p.conf.TagName = "json"
	}
}

// Table returns a new TableParser to parse a table
// It parses a table from an io.Reader
func Table() *TableParser {
	return &TableParser{}
}

func (p *TableParser) Unmarshal(b []byte, obj any) error {
	p.lazyInit()

	p.conf.Result = obj
	dec, err := mapstructure.NewDecoder(p.conf)
	if err != nil {
		return errors.Wrap(err, "failed to create decoder")
	}

	// get the maximum length of the line
	scanner := bufio.NewScanner(bytes.NewReader(b))
	lines := make([]string, 0)
	for i := 0; scanner.Scan(); i++ {
		if p.filter(i, scanner.Text()) {
			lines = append(lines, scanner.Text())
		}
	}
	if scanner.Err() != nil {
		return errors.Wrap(scanner.Err(), "failed to scan")
	}
	if len(lines) == 0 {
		return nil
	}

	headerTxt := p.headerTxt
	if len(headerTxt) == 0 {
		headerTxt, lines = lines[0], lines[1:]
		if len(lines) == 0 {
			return nil
		}
	}
	header := TableHeader(headerTxt)
	headerIndex, err := header.Index(p.separatorFunc, lines)
	if err != nil {
		return errors.Wrap(err, "failed to parse header")
	}

	// parse the table
	ret := make([]map[string]any, 0, len(lines))
	for _, line := range lines {
		item := make(map[string]any, len(headerIndex))
		for _, h := range headerIndex {
			k := p.keyTransformFunc(headerTxt[h[0]:min(h[1], len(headerTxt))])
			v := p.valTransformFunc(line[h[0]:min(h[1], len(line))])
			item[k] = v
		}
		ret = append(ret, item)
	}

	err = dec.Decode(ret)
	if err != nil {
		return errors.Wrap(err, "failed to decode")
	}
	return nil
}
