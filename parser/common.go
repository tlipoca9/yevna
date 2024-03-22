package parser

import (
	"io"

	"github.com/cockroachdb/errors"
)

type commonParser struct {
	unmarshal func([]byte, any) error
	dataType  DataType
}

func (p *commonParser) Parse(r io.Reader) (any, error) {
	buf, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	switch p.dataType {
	case Object:
		obj := make(map[string]any)
		err = p.unmarshal(buf, &obj)
		if err != nil {
			return nil, err
		}
		return obj, nil
	case Array:
		var arr []any
		err = p.unmarshal(buf, &arr)
		if err != nil {
			return nil, err
		}
		return arr, nil
	default:
		return nil, errors.New("invalid data type")
	}
}

type CommonParserBuilder struct {
	data commonParser
}

func (b *CommonParserBuilder) Unmarshal(f func([]byte, any) error) *CommonParserBuilder {
	b.data.unmarshal = f
	return b
}

func (b *CommonParserBuilder) DataType(t DataType) *CommonParserBuilder {
	b.data.dataType = t
	return b
}

func (b *CommonParserBuilder) Build() Parser {
	return &b.data
}

func (b *CommonParserBuilder) Parse(r io.Reader) (any, error) {
	return b.Build().Parse(r)
}

func Common() *CommonParserBuilder {
	return &CommonParserBuilder{}
}
