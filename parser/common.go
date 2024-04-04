package parser

import (
	"github.com/goccy/go-json"
	"github.com/goccy/go-yaml"
	"github.com/pelletier/go-toml/v2"
)

// JSON is the shortcut for Func(json.Unmarshal)
func JSON() Parser {
	return Func(json.Unmarshal)
}

// YAML is the shortcut for Func(yaml.Unmarshal)
func YAML() Parser {
	return Func(yaml.Unmarshal)
}

// TOML is the shortcut for Func(toml.Unmarshal)
func TOML() Parser {
	return Func(toml.Unmarshal)
}
