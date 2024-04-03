package parser

import (
	"github.com/cockroachdb/errors"
	"github.com/go-viper/mapstructure/v2"
	"github.com/joho/godotenv"
)

type DotenvParser struct {
	conf *mapstructure.DecoderConfig
}

func (p *DotenvParser) WithDecoderConfig(conf *mapstructure.DecoderConfig) *DotenvParser {
	p.conf = conf
	return p
}

func (p *DotenvParser) lazyInit() {
	if p.conf == nil {
		p.conf = &mapstructure.DecoderConfig{TagName: "dotenv"}
	}
	if p.conf.TagName == "" {
		p.conf.TagName = "dotenv"
	}
}

// Dotenv returns a new DotenvParser
func Dotenv() *DotenvParser {
	return &DotenvParser{}
}

func (p *DotenvParser) Unmarshal(b []byte, v any) error {
	p.lazyInit()

	p.conf.Result = v
	dec, err := mapstructure.NewDecoder(p.conf)
	if err != nil {
		return errors.Wrapf(err, "create decoder failed")
	}
	raw, err := godotenv.UnmarshalBytes(b)
	if err != nil {
		return errors.Wrapf(err, "unmarshal failed")
	}
	err = dec.Decode(raw)
	if err != nil {
		return errors.Wrapf(err, "decode failed")
	}
	return nil
}
