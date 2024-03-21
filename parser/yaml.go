package parser

import "github.com/goccy/go-yaml"

func YAMLObject() Parser {
	return YAML(Object)
}

func YAMLArray() Parser {
	return YAML(Array)
}

func YAML(t ...DataType) Parser {
	dt := Object
	if len(t) > 0 {
		dt = t[0]
	}
	return &CommonParser{
		Unmarshal: yaml.Unmarshal,
		DataType:  dt,
	}
}
