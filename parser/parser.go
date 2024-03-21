package parser

import "io"

type Parser interface {
	Parse(r io.Reader) (any, error)
}

type DataType int

const (
	Object DataType = iota + 1
	Array
)
