package parser

import (
	"io"

	"github.com/goccy/go-yaml"
)

type YAMLParserBuilder struct {
	data commonParser
}

func (b *YAMLParserBuilder) DataType(t DataType) *YAMLParserBuilder {
	b.data.dataType = t
	return b
}

func (b *YAMLParserBuilder) Build() Parser {
	if b.data.unmarshal == nil {
		b.data.unmarshal = yaml.Unmarshal
	}
	if b.data.dataType == 0 {
		b.data.dataType = Object
	}
	return &b.data
}

func (b *YAMLParserBuilder) Parse(r io.Reader) (any, error) {
	return b.Build().Parse(r)
}

func YAML() *YAMLParserBuilder {
	return &YAMLParserBuilder{}
}
