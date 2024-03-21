package parser

import (
	"github.com/goccy/go-json"
)

func JSONObject() Parser {
	return JSON(Object)
}

func JSONArray() Parser {
	return JSON(Array)
}

func JSON(t ...DataType) Parser {
	dt := Object
	if len(t) > 0 {
		dt = t[0]
	}
	return &CommonParser{
		Unmarshal: json.Unmarshal,
		DataType:  dt,
	}
}
