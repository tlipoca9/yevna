package parser

import (
	"errors"
	"io"
)

type CommonParser struct {
	Unmarshal func([]byte, any) error
	DataType  DataType
}

func (p *CommonParser) Parse(r io.Reader) (any, error) {
	buf, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	switch p.DataType {
	case Object:
		obj := make(map[string]any)
		err = p.Unmarshal(buf, &obj)
		if err != nil {
			return nil, err
		}
		return obj, nil
	case Array:
		var arr []any
		err = p.Unmarshal(buf, &arr)
		if err != nil {
			return nil, err
		}
		return arr, nil
	default:
		return nil, errors.New("invalid data type")
	}
}
