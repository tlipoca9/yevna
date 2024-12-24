package parser

import (
	"bytes"
	"encoding/csv"
	"io"
	"slices"

	"github.com/cockroachdb/errors"
	"github.com/go-viper/mapstructure/v2"
)

type CsvParser struct {
	conf    *mapstructure.DecoderConfig
	headers []string
}

func (p *CsvParser) WithDecoderConfig(conf *mapstructure.DecoderConfig) *CsvParser {
	p.conf = conf
	return p
}

func (p *CsvParser) WithHeaders(headers ...string) *CsvParser {
	p.headers = headers
	return p
}

func (p *CsvParser) lazyInit() {
	if p.conf == nil {
		p.conf = &mapstructure.DecoderConfig{TagName: "json"}
	}
	if p.conf.TagName == "" {
		p.conf.TagName = "json"
	}
}

// CSV returns a new CsvParser
func CSV() *CsvParser {
	return &CsvParser{}
}

func (p *CsvParser) Unmarshal(b []byte, v any) error {
	p.lazyInit()

	p.conf.Result = v
	dec, err := mapstructure.NewDecoder(p.conf)
	if err != nil {
		return errors.Wrapf(err, "create decoder failed")
	}

	var (
		results []any
		record  []string
	)
	r := csv.NewReader(bytes.NewBuffer(b))
	r.ReuseRecord = true
	if len(p.headers) != 0 {
		r.FieldsPerRecord = len(p.headers)
	} else {
		record, err = r.Read()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return errors.Wrapf(err, "read header failed")
		}
		p.headers = slices.Clone(record)
	}
	for record, err = r.Read(); err == nil; record, err = r.Read() {
		m := make(map[string]string)
		for i, key := range p.headers {
			m[key] = record[i]
		}
		results = append(results, m)
	}
	if !errors.Is(err, io.EOF) {
		return errors.Wrapf(err, "read record failed")
	}

	err = dec.Decode(results)
	if err != nil {
		return errors.Wrapf(err, "decode failed")
	}

	return nil
}
