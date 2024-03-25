package parser

import (
	"io"

	"github.com/joho/godotenv"
)

type DotenvParser struct{}

func (DotenvParser) Parse(r io.Reader) (any, error) {
	return godotenv.Parse(r)
}

// Dotenv returns a new DotenvParser
func Dotenv() *DotenvParser {
	return &DotenvParser{}
}
