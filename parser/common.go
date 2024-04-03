package parser

import (
	"github.com/goccy/go-json"
	"github.com/goccy/go-yaml"
)

type CommonParserOptions struct {
	UnmarshalFunc func([]byte, any) error
}

type CommonParser struct {
	opt CommonParserOptions
}

func (p *CommonParser) WithUnmarshalFunc(f func([]byte, any) error) *CommonParser {
	p.opt.UnmarshalFunc = f
	return p
}

func (p *CommonParser) Unmarshal(b []byte, v any) error {
	return p.opt.UnmarshalFunc(b, v)
}

// Common returns a new CommonParser
func Common() *CommonParser {
	return &CommonParser{}
}

// JSON is the shortcut for Common().WithUnmarshalFunc(json.Unmarshal)
func JSON() *CommonParser {
	return Common().WithUnmarshalFunc(json.Unmarshal)
}

// YAML is the shortcut for Common().WithUnmarshalFunc(yaml.Unmarshal)
func YAML() *CommonParser {
	return Common().WithUnmarshalFunc(yaml.Unmarshal)
}
