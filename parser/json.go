package parser

import (
	"io"

	"github.com/goccy/go-json"
)

type JSONParserBuilder struct {
	data commonParser
}

func (b *JSONParserBuilder) DataType(t DataType) *JSONParserBuilder {
	b.data.dataType = t
	return b
}

func (b *JSONParserBuilder) Build() Parser {
	if b.data.unmarshal == nil {
		b.data.unmarshal = json.Unmarshal
	}
	if b.data.dataType == 0 {
		b.data.dataType = Object
	}
	return &b.data
}

func (b *JSONParserBuilder) Parse(r io.Reader) (any, error) {
	return b.Build().Parse(r)
}

func JSON() *JSONParserBuilder {
	return &JSONParserBuilder{}
}
