package parser

import (
	"github.com/joho/godotenv"
	"io"
)

type dotenvParser struct{}

func Dotenv() Parser {
	return &dotenvParser{}
}

func (dotenvParser) Parse(r io.Reader) (any, error) {
	return godotenv.Parse(r)
}
