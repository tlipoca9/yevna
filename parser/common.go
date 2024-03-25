package parser

import (
	"io"

	"github.com/cockroachdb/errors"
	"github.com/goccy/go-json"
	"github.com/goccy/go-yaml"
)

type CommonParserOptions struct {
	UnmarshalFunc func([]byte, any) error
	DataType      DataType
}

type CommonParser struct {
	opt CommonParserOptions
}

func (p *CommonParser) WithUnmarshalFunc(f func([]byte, any) error) *CommonParser {
	p.opt.UnmarshalFunc = f
	return p
}

func (p *CommonParser) WithDataType(t DataType) *CommonParser {
	p.opt.DataType = t
	return p
}

func (p *CommonParser) lazyInit() {
	if p.opt.UnmarshalFunc == nil {
		p.opt.UnmarshalFunc = json.Unmarshal
	}
	if p.opt.DataType == 0 {
		p.opt.DataType = Object
	}
}

func (p *CommonParser) Parse(r io.Reader) (any, error) {
	p.lazyInit()

	buf, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	switch p.opt.DataType {
	case Object:
		obj := make(map[string]any)
		err = p.opt.UnmarshalFunc(buf, &obj)
		if err != nil {
			return nil, err
		}
		return obj, nil
	case Array:
		var arr []any
		err = p.opt.UnmarshalFunc(buf, &arr)
		if err != nil {
			return nil, err
		}
		return arr, nil
	default:
		return nil, errors.New("invalid opt type")
	}
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
