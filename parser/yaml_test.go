package parser_test

import (
	"strings"

	"github.com/tlipoca9/yevna/parser"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("YAMLParser", func() {
	Context("Object", func() {
		p := parser.YAML()
		When("input is empty", func() {
			It("return empty object", func() {
				got, err := p.Parse(strings.NewReader(""))
				Expect(err).To(BeNil())
				Expect(got).To(BeEmpty())
			})
		})

		When("input is simple", func() {
			It("return expected object", func() {
				got, err := p.Parse(strings.NewReader("FOO: BAR\n"))
				Expect(err).To(BeNil())
				Expect(got).To(Equal(map[string]any{"FOO": "BAR"}))
			})
		})

		When("input is multiline", func() {
			It("return expected object", func() {
				got, err := p.Parse(strings.NewReader("FOO: BAR\nBAZ: QUX\n"))
				Expect(err).To(BeNil())
				Expect(got).To(Equal(map[string]any{"FOO": "BAR", "BAZ": "QUX"}))
			})
		})
	})

	Context("Array", func() {
		p := parser.YAML().DataType(parser.Array)
		When("input is empty", func() {
			It("return empty object", func() {
				got, err := p.Parse(strings.NewReader("[]"))
				Expect(err).To(BeNil())
				Expect(got).To(BeEmpty())
			})
		})

		When("input is simple", func() {
			It("return expected object", func() {
				got, err := p.Parse(strings.NewReader("- FOO\n"))
				Expect(err).To(BeNil())
				Expect(got).To(Equal([]any{"FOO"}))
			})
		})

		When("input is multiline", func() {
			It("return expected object", func() {
				got, err := p.Parse(strings.NewReader("- FOO\n- BAR\n"))
				Expect(err).To(BeNil())
				Expect(got).To(Equal([]any{"FOO", "BAR"}))
			})
		})
	})
})
