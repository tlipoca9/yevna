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

func (h TableHeader) Index(isSep func(rune) bool, lines []string) ([][2]int, error) {
	ret := make([][2]int, 0)
	sepMap := make([]bool, len(h)+1)
	for i, c := range h {
		if !isSep(c) {
			sepMap[i] = true
		}
	}

	for _, line := range lines {
		if len(sepMap) < len(line)+1 {
			sepMap = append(sepMap, make([]bool, len(line)+1-len(sepMap))...)
		}
		for i, c := range line {
			if !sepMap[i] && !isSep(c) {
				sepMap[i] = true
			}
		}
	}

	i := 0
	for idx, v := range sepMap {
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
	conf      *mapstructure.DecoderConfig
	sepFunc   func(rune) bool
	filter    func(i int, line string) bool
	cb        func(k, v string) (string, string)
	headerTxt string
}

func (p *TableParser) WithDecoderConfig(conf *mapstructure.DecoderConfig) *TableParser {
	p.conf = conf
	return p
}

func (p *TableParser) WithSepFunc(f func(rune) bool) *TableParser {
	p.sepFunc = f
	return p
}

func (p *TableParser) WithFilter(f func(i int, line string) bool) *TableParser {
	p.filter = f
	return p
}

func (p *TableParser) WithCallback(f func(k, v string) (string, string)) *TableParser {
	p.cb = f
	return p
}

func (p *TableParser) WithHeader(header string) *TableParser {
	p.headerTxt = header
	return p
}

func (p *TableParser) lazyInit() {
	if p.sepFunc == nil {
		p.sepFunc = unicode.IsSpace
	}
	if p.filter == nil {
		p.filter = func(_ int, _ string) bool { return true }
	}
	if p.cb == nil {
		p.cb = func(k, v string) (string, string) {
			return strings.TrimFunc(k, p.sepFunc), strings.TrimFunc(v, p.sepFunc)
		}
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

	// filter the lines
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

	// parse the header
	headerTxt := p.headerTxt
	if len(headerTxt) == 0 {
		headerTxt, lines = lines[0], lines[1:]
		if len(lines) == 0 {
			return nil
		}
	}
	header := TableHeader(headerTxt)
	headerIndex, err := header.Index(p.sepFunc, lines)
	if err != nil {
		return errors.Wrap(err, "failed to parse header")
	}

	// parse the table
	ret := make([]map[string]any, 0, len(lines))
	for _, line := range lines {
		item := make(map[string]any, len(headerIndex))
		for _, h := range headerIndex {
			k := headerTxt[h[0]:min(h[1], len(headerTxt))]
			v := line[h[0]:min(h[1], len(line))]
			k, v = p.cb(k, v)
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
