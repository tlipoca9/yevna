package parser_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/tlipoca9/yevna/parser"
)

func TestTable(t *testing.T) {
	cases := []struct {
		name     string
		input    string
		expected any
		err      bool
	}{
		{
			name:     "empty",
			input:    "",
			expected: nil,
		},
		{
			name: "eza -l --header",
			input: `Permissions Size User Date Modified Name
drwxr-xr-x     - foo  21 Mar 09:58  cmd
drwxr-xr-x     - foo  21 Mar 09:42  cmdx
drwxr-xr-x     - foo  21 Mar 10:03  execx
.rw-r--r--  1.0k foo  21 Mar 10:11  go.mod
.rw-r--r--   12k foo  21 Mar 10:11  go.sum
.rw-r--r--   342 foo  21 Mar 09:59  main.go
.rw-r--r--   132 foo  21 Mar 09:42  Makefile
drwxr-xr-x     - foo  21 Mar 10:50  parser
`,
			expected: []map[string]any{
				{
					"Permissions":   "drwxr-xr-x",
					"Size":          "-",
					"User":          "foo",
					"Date Modified": "21 Mar 09:58",
					"Name":          "cmd",
				},
				{
					"Permissions":   "drwxr-xr-x",
					"Size":          "-",
					"User":          "foo",
					"Date Modified": "21 Mar 09:42",
					"Name":          "cmdx",
				},
				{
					"Permissions":   "drwxr-xr-x",
					"Size":          "-",
					"User":          "foo",
					"Date Modified": "21 Mar 10:03",
					"Name":          "execx",
				},
				{
					"Permissions":   ".rw-r--r--",
					"Size":          "1.0k",
					"User":          "foo",
					"Date Modified": "21 Mar 10:11",
					"Name":          "go.mod",
				},
				{
					"Permissions":   ".rw-r--r--",
					"Size":          "12k",
					"User":          "foo",
					"Date Modified": "21 Mar 10:11",
					"Name":          "go.sum",
				},
				{
					"Permissions":   ".rw-r--r--",
					"Size":          "342",
					"User":          "foo",
					"Date Modified": "21 Mar 09:59",
					"Name":          "main.go",
				},
				{
					"Permissions":   ".rw-r--r--",
					"Size":          "132",
					"User":          "foo",
					"Date Modified": "21 Mar 09:42",
					"Name":          "Makefile",
				},
				{
					"Permissions":   "drwxr-xr-x",
					"Size":          "-",
					"User":          "foo",
					"Date Modified": "21 Mar 10:50",
					"Name":          "parser",
				},
			},
		},
		{
			name: "kubectl get pods",
			input: `NAME                                                              READY   STATUS                       RESTARTS            AGE
foofoofoofoofoofoofoofoo-74bc6cbf96-wgxn8                         1/1     Running                      0                   2d13h
foofoofoofoofoofoofoofoofoofoof-66b86b4ccf-7mnzd                  0/1     CrashLoopBackOff             495 (4m13s ago)     42h
foofoofoofoofoofoofoofoofoofoof-66b86b4ccf-lckbm                  0/1     CrashLoopBackOff             9063 (3m4s ago)     39d
foofoofoofoofoofoofoofoofoofoof-785fc5694b-82cll                  0/1     CrashLoopBackOff             734 (52s ago)       2d15h
foofoofoofoofoofoofoofoofoofoof-785fc5694b-r766g                  0/1     CrashLoopBackOff             472 (4m53s ago)     40h
foofoof-7c65b8458c-wmxc2                                          1/1     Running                      0                   40h
foofoofoofoofoofoo-6ff6f5c957-9rmtv                               1/1     Running                      0                   40h
foofoofoofoofoofoofoofoofoofoof-28516525-sfb8g                    0/1     Completed                    0                   13m
foofoofoofoofoofoofoofoofoofoof-28516530-sxr74                    0/1     Completed                    0                   8m9s
`,
			expected: []map[string]any{
				{
					"NAME":     "foofoofoofoofoofoofoofoo-74bc6cbf96-wgxn8",
					"READY":    "1/1",
					"STATUS":   "Running",
					"RESTARTS": "0",
					"AGE":      "2d13h",
				},
				{
					"NAME":     "foofoofoofoofoofoofoofoofoofoof-66b86b4ccf-7mnzd",
					"READY":    "0/1",
					"STATUS":   "CrashLoopBackOff",
					"RESTARTS": "495 (4m13s ago)",
					"AGE":      "42h",
				},
				{
					"NAME":     "foofoofoofoofoofoofoofoofoofoof-66b86b4ccf-lckbm",
					"READY":    "0/1",
					"STATUS":   "CrashLoopBackOff",
					"RESTARTS": "9063 (3m4s ago)",
					"AGE":      "39d",
				},
				{
					"NAME":     "foofoofoofoofoofoofoofoofoofoof-785fc5694b-82cll",
					"READY":    "0/1",
					"STATUS":   "CrashLoopBackOff",
					"RESTARTS": "734 (52s ago)",
					"AGE":      "2d15h",
				},
				{
					"NAME":     "foofoofoofoofoofoofoofoofoofoof-785fc5694b-r766g",
					"READY":    "0/1",
					"STATUS":   "CrashLoopBackOff",
					"RESTARTS": "472 (4m53s ago)",
					"AGE":      "40h",
				},
				{
					"NAME":     "foofoof-7c65b8458c-wmxc2",
					"READY":    "1/1",
					"STATUS":   "Running",
					"RESTARTS": "0",
					"AGE":      "40h",
				},
				{
					"NAME":     "foofoofoofoofoofoo-6ff6f5c957-9rmtv",
					"READY":    "1/1",
					"STATUS":   "Running",
					"RESTARTS": "0",
					"AGE":      "40h",
				},
				{
					"NAME":     "foofoofoofoofoofoofoofoofoofoof-28516525-sfb8g",
					"READY":    "0/1",
					"STATUS":   "Completed",
					"RESTARTS": "0",
					"AGE":      "13m",
				},
				{
					"NAME":     "foofoofoofoofoofoofoofoofoofoof-28516530-sxr74",
					"READY":    "0/1",
					"STATUS":   "Completed",
					"RESTARTS": "0",
					"AGE":      "8m9s",
				},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			p := parser.Table()
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

func benchmarkTable(i int, b *testing.B) {
	input := "Permissions Size User Date Modified Name\n"
	input += strings.Repeat(`drwxr-xr-x     - foo  21 Mar 09:42  cmdx
drwxr-xr-x     - foo  21 Mar 10:03  execx
.rw-r--r--  1.0k foo  21 Mar 10:11  go.mod
.rw-r--r--   12k foo  21 Mar 10:11  go.sum
.rw-r--r--   342 foo  21 Mar 09:59  main.go
.rw-r--r--   132 foo  21 Mar 09:42  Makefile
drwxr-xr-x     - foo  21 Mar 10:50  parser
`, i)
	p := parser.Table()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := p.Parse(strings.NewReader(input))
		if err != nil {
			b.Fatalf("unexpected error: %v", err)
		}
	}
}

func BenchmarkTable_10(b *testing.B)    { benchmarkTable(10, b) }
func BenchmarkTable_100(b *testing.B)   { benchmarkTable(100, b) }
func BenchmarkTable_1000(b *testing.B)  { benchmarkTable(1000, b) }
func BenchmarkTable_10000(b *testing.B) { benchmarkTable(10000, b) }

func benchmarkTable2(i int, b *testing.B) {
	input := "NAME                                                              READY   STATUS                       RESTARTS            AGE\n"
	input += strings.Repeat(`foofoofoofoofoofoofoofoofoofoof-66b86b4ccf-7mnzd                  0/1     CrashLoopBackOff             495 (4m13s ago)     42h
foofoofoofoofoofoofoofoofoofoof-66b86b4ccf-lckbm                  0/1     CrashLoopBackOff             9063 (3m4s ago)     39d
foofoofoofoofoofoofoofoofoofoof-785fc5694b-82cll                  0/1     CrashLoopBackOff             734 (52s ago)       2d15h
foofoofoofoofoofoofoofoofoofoof-785fc5694b-r766g                  0/1     CrashLoopBackOff             472 (4m53s ago)     40h
foofoof-7c65b8458c-wmxc2                                          1/1     Running                      0                   40h
foofoofoofoofoofoo-6ff6f5c957-9rmtv                               1/1     Running                      0                   40h
foofoofoofoofoofoofoofoofoofoof-28516525-sfb8g                    0/1     Completed                    0                   13m
foofoofoofoofoofoofoofoofoofoof-28516530-sxr74                    0/1     Completed                    0                   8m9s
`, i)
	p := parser.Table()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := p.Parse(strings.NewReader(input))
		if err != nil {
			b.Fatalf("unexpected error: %v", err)
		}
	}
}

func BenchmarkTable2_10(b *testing.B)    { benchmarkTable2(10, b) }
func BenchmarkTable2_100(b *testing.B)   { benchmarkTable2(100, b) }
func BenchmarkTable2_1000(b *testing.B)  { benchmarkTable2(1000, b) }
func BenchmarkTable2_10000(b *testing.B) { benchmarkTable2(10000, b) }
