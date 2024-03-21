package parser

import (
	"io"

	"github.com/joho/godotenv"
)

type dotenvParser struct{}

func Dotenv() Parser {
	return &dotenvParser{}
}

func (dotenvParser) Parse(r io.Reader) (any, error) {
	return godotenv.Parse(r)
}
