package parser_test

import (
	"strings"
	"time"

	"github.com/onsi/gomega/gmeasure"

	"github.com/tlipoca9/yevna/parser"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("TableParser", func() {
	p := parser.Table()
	var got []map[string]any

	BeforeEach(func() {
		got = nil
	})

	When("input is empty", func() {
		It("return empty object", func() {
			buf := []byte("")

			err := p.Unmarshal(buf, &got)
			Expect(err).To(BeNil())
			Expect(got).To(BeEmpty())
			Expect(len(got)).To(Equal(cap(got)))
		})
	})

	When("input is eza -l --header", func() {
		It("return expected object", func() {
			buf := []byte(`
Permissions Size User Date Modified Name
drwxr-xr-x     - foo  21 Mar 09:58  cmd
drwxr-xr-x     - foo  21 Mar 09:42  cmdx
drwxr-xr-x     - foo  21 Mar 10:03  execx
.rw-r--r--  1.0k foo  21 Mar 10:11  go.mod
.rw-r--r--   12k foo  21 Mar 10:11  go.sum
.rw-r--r--   342 foo  21 Mar 09:59  main.go
.rw-r--r--   132 foo  21 Mar 09:42  Makefile
drwxr-xr-x     - foo  21 Mar 10:50  parser
`[1:])

			err := p.Unmarshal(buf, &got)
			Expect(err).To(BeNil())
			Expect(got).To(Equal([]map[string]any{
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
			}))
			Expect(len(got)).To(Equal(cap(got)))
		})
	})

	When("input is kubectl get pods", func() {
		It("return expected object", func() {
			buf := []byte(`
NAME                                                              READY   STATUS                       RESTARTS            AGE
foofoofoofoofoofoofoofoo-74bc6cbf96-wgxn8                         1/1     Running                      0                   2d13h
foofoofoofoofoofoofoofoofoofoof-66b86b4ccf-7mnzd                  0/1     CrashLoopBackOff             495 (4m13s ago)     42h
foofoofoofoofoofoofoofoofoofoof-66b86b4ccf-lckbm                  0/1     CrashLoopBackOff             9063 (3m4s ago)     39d
foofoofoofoofoofoofoofoofoofoof-785fc5694b-82cll                  0/1     CrashLoopBackOff             734 (52s ago)       2d15h
foofoofoofoofoofoofoofoofoofoof-785fc5694b-r766g                  0/1     CrashLoopBackOff             472 (4m53s ago)     40h
foofoof-7c65b8458c-wmxc2                                          1/1     Running                      0                   40h
foofoofoofoofoofoo-6ff6f5c957-9rmtv                               1/1     Running                      0                   40h
foofoofoofoofoofoofoofoofoofoof-28516525-sfb8g                    0/1     Completed                    0                   13m
foofoofoofoofoofoofoofoofoofoof-28516530-sxr74                    0/1     Completed                    0                   8m9s
`[1:])

			err := p.Unmarshal(buf, &got)
			Expect(err).To(BeNil())
			Expect(got).To(Equal([]map[string]any{
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
			}))
			Expect(len(got)).To(Equal(cap(got)))
		})
	})
})

func generateTableBenchmarkInput(i int) []byte {
	input := "Permissions Size User Date Modified Name\n"
	input += strings.Repeat(`
drwxr-xr-x     - foo  21 Mar 09:42  cmdx
drwxr-xr-x     - foo  21 Mar 10:03  execx
.rw-r--r--  1.0k foo  21 Mar 10:11  go.mod
.rw-r--r--   12k foo  21 Mar 10:11  go.sum
.rw-r--r--   342 foo  21 Mar 09:59  main.go
.rw-r--r--   132 foo  21 Mar 09:42  Makefile
drwxr-xr-x     - foo  21 Mar 10:50  parser
`[1:], i)
	return []byte(input)
}

var _ = Describe("Benchmark TableParser", func() {
	It("benchmark 1e2", Serial, Label("parser", "table", "benchmark"), func() {
		experiment := gmeasure.NewExperiment("benchmark_table_1e2")
		AddReportEntry(experiment.Name, experiment)
		experiment.Sample(func(_ int) {
			r, p := generateTableBenchmarkInput(1e2), parser.Table()
			got := make([]map[string]any, 0)
			experiment.MeasureDuration("unmarshal", func() {
				err := p.Unmarshal(r, &got)
				Expect(err).To(BeNil())
			})
		}, gmeasure.SamplingConfig{N: 1e3, Duration: time.Second})
	})

	It("benchmark 1e3", Serial, Label("parser", "table", "benchmark"), func() {
		experiment := gmeasure.NewExperiment("benchmark_table_1e3")
		AddReportEntry(experiment.Name, experiment)
		experiment.Sample(func(_ int) {
			r, p := generateTableBenchmarkInput(1e3), parser.Table()
			got := make([]map[string]any, 0)
			experiment.MeasureDuration("unmarshal", func() {
				err := p.Unmarshal(r, &got)
				Expect(err).To(BeNil())
			})
		}, gmeasure.SamplingConfig{N: 1e3, Duration: time.Second})
	})

	It("benchmark 1e4", Serial, Label("parser", "table", "benchmark"), func() {
		experiment := gmeasure.NewExperiment("benchmark_table_1e4")
		AddReportEntry(experiment.Name, experiment)
		experiment.Sample(func(_ int) {
			r, p := generateTableBenchmarkInput(1e4), parser.Table()
			got := make([]map[string]any, 0)
			experiment.MeasureDuration("unmarshal", func() {
				err := p.Unmarshal(r, &got)
				Expect(err).To(BeNil())
			})
		}, gmeasure.SamplingConfig{N: 1e3, Duration: time.Second})
	})
})
