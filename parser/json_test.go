package parser_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/tlipoca9/yevna/parser"
)

func TestJSONObject(t *testing.T) {
	cases := []struct {
		name     string
		input    string
		expected any
		err      bool
	}{
		{
			name:     "empty",
			input:    "{}",
			expected: map[string]any{},
		},
		{
			name:     "simple",
			input:    `{"FOO":"BAR"}`,
			expected: map[string]any{"FOO": "BAR"},
		},
		{
			name:     "multiline",
			input:    `{"FOO":"BAR","BAZ":"QUX"}`,
			expected: map[string]any{"FOO": "BAR", "BAZ": "QUX"},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			p := parser.JSONObject()
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

func TestJSONArray(t *testing.T) {
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
			input:    `["FOO"]`,
			expected: []any{"FOO"},
		},
		{
			name:     "multiline",
			input:    `["FOO","BAR"]`,
			expected: []any{"FOO", "BAR"},
		},
		{
			name:     "nested",
			input:    `[["FOO","BAR"],["BAZ","QUX"]]`,
			expected: []any{[]any{"FOO", "BAR"}, []any{"BAZ", "QUX"}},
		},
		{
			name:     "mixed",
			input:    `["FOO",{"BAZ":"QUX"}]`,
			expected: []any{"FOO", map[string]any{"BAZ": "QUX"}},
		},
		{
			name:     "multi_object",
			input:    `[{"FOO":"BAR"},{"BAZ":"QUX"}]`,
			expected: []any{map[string]any{"FOO": "BAR"}, map[string]any{"BAZ": "QUX"}},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			p := parser.JSONArray()
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
