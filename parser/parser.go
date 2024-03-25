package parser

import "io"

type ParseFunc func(r io.Reader) (any, error)

func (f ParseFunc) Parse(r io.Reader) (any, error) {
	return f(r)
}

type Parser interface {
	Parse(r io.Reader) (any, error)
}

type DataType int

const (
	Object DataType = iota + 1
	Array
)
