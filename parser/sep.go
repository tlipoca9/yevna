package parser

import (
	"bufio"
	"bytes"

	"github.com/cockroachdb/errors"
	"github.com/go-viper/mapstructure/v2"
)

type SepParser struct {
	conf      *mapstructure.DecoderConfig
	splitFunc bufio.SplitFunc
	filter    func(token string) bool
}

func (p *SepParser) SplitFunc(f bufio.SplitFunc) *SepParser {
	p.splitFunc = f
	return p
}

func (p *SepParser) Filter(f func(token string) bool) *SepParser {
	p.filter = f
	return p
}

func (p *SepParser) WithDecoderConfig(conf *mapstructure.DecoderConfig) *SepParser {
	p.conf = conf
	return p
}

func (p *SepParser) lazyInit() {
	if p.splitFunc == nil {
		p.splitFunc = bufio.ScanWords
	}
	if p.filter == nil {
		p.filter = func(_ string) bool { return true }
	}
	if p.conf == nil {
		p.conf = &mapstructure.DecoderConfig{TagName: "json"}
	}
	if p.conf.TagName == "" {
		p.conf.TagName = "json"
	}
}

// Sep returns a new SepParser
func Sep() *SepParser {
	return &SepParser{}
}

// Line is the shortcut for Sep().splitFunc(bufio.ScanLines)
func Line() *SepParser {
	return Sep().SplitFunc(bufio.ScanLines)
}

func (p *SepParser) Unmarshal(b []byte, v any) error {
	p.lazyInit()

	p.conf.Result = v
	dec, err := mapstructure.NewDecoder(p.conf)
	if err != nil {
		return errors.Wrapf(err, "create decoder failed")
	}

	scanner := bufio.NewScanner(bytes.NewReader(b))
	scanner.Split(p.splitFunc)
	var tokens []string
	for scanner.Scan() {
		token := scanner.Text()
		if p.filter(token) {
			tokens = append(tokens, token)
		}
	}
	if scanner.Err() != nil {
		return errors.Wrapf(scanner.Err(), "scan failed")
	}

	err = dec.Decode(tokens)
	if err != nil {
		return errors.Wrapf(err, "decode failed")
	}

	return nil
}
