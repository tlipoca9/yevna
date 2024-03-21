package parser_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/tlipoca9/yevna/parser"
)

func TestYAMLObject(t *testing.T) {
	cases := []struct {
		name     string
		input    string
		expected any
		err      bool
	}{
		{
			name:     "empty",
			input:    "",
			expected: map[string]any{},
		},
		{
			name:     "simple",
			input:    "FOO: BAR\n",
			expected: map[string]any{"FOO": "BAR"},
		},
		{
			name:     "multiline",
			input:    "FOO: BAR\nBAZ: QUX\n",
			expected: map[string]any{"FOO": "BAR", "BAZ": "QUX"},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			p := parser.YAMLObject()
			got, err := p.Parse(strings.NewReader(c.input))
			if c.err && err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !c.err && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(got, c.expected) {
				t.Fatalf("expected: %v, got: %v", c.expected, got)
			}
		})
	}
}

func TestYAMLArray(t *testing.T) {
	cases := []struct {
		name     string
		input    string
		expected any
		err      bool
	}{
		{
			name:     "empty",
			input:    "[]",
			expected: []any{},
		},
		{
			name:     "simple",
			input:    "- FOO\n",
			expected: []any{"FOO"},
		},
		{
			name:     "multiline",
			input:    "- FOO\n- BAR\n",
			expected: []any{"FOO", "BAR"},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			p := parser.YAMLArray()
			got, err := p.Parse(strings.NewReader(c.input))
			if c.err && err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !c.err && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(got, c.expected) {
				t.Fatalf("expected: %v, got: %v", c.expected, got)
			}
		})
	}
}
