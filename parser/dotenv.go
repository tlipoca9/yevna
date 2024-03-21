package parser

import (
	"io"

	"github.com/joho/godotenv"
)

type dotenvParser struct{}

func (dotenvParser) Parse(r io.Reader) (any, error) {
	return godotenv.Parse(r)
}

type DotenvParserBuilder struct {
	data dotenvParser
}

func (b *DotenvParserBuilder) Build() Parser {
	return &b.data
}

func (b *DotenvParserBuilder) Parse(r io.Reader) (any, error) {
	return b.Build().Parse(r)
}

func Dotenv() *DotenvParserBuilder {
	return &DotenvParserBuilder{}
}
