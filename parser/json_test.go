package parser_test

import (
	"strings"

	"github.com/tlipoca9/yevna/parser"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("JSONParser", func() {
	Context("Object", func() {
		p := parser.JSON()
		When("input is empty", func() {
			It("return empty object", func() {
				got, err := p.Parse(strings.NewReader("{}"))
				Expect(err).To(BeNil())
				Expect(got).To(BeEmpty())
			})
		})

		When("input is simple", func() {
			It("return expected object", func() {
				got, err := p.Parse(strings.NewReader(`{"FOO":"BAR"}`))
				Expect(err).To(BeNil())
				Expect(got).To(Equal(map[string]any{"FOO": "BAR"}))
			})
		})

		When("input is multiline", func() {
			It("return expected object", func() {
				got, err := p.Parse(strings.NewReader(`{"FOO":"BAR","BAZ":"QUX"}`))
				Expect(err).To(BeNil())
				Expect(got).To(Equal(map[string]any{"FOO": "BAR", "BAZ": "QUX"}))
			})
		})
	})

	Context("Array", func() {
		p := parser.JSON().DataType(parser.Array)
		When("input is empty", func() {
			It("return empty object", func() {
				got, err := p.Parse(strings.NewReader("[]"))
				Expect(err).To(BeNil())
				Expect(got).To(BeEmpty())
			})
		})

		When("input is simple", func() {
			It("return expected object", func() {
				got, err := p.Parse(strings.NewReader(`["FOO"]`))
				Expect(err).To(BeNil())
				Expect(got).To(Equal([]any{"FOO"}))
			})
		})

		When("input is multiline", func() {
			It("return expected object", func() {
				got, err := p.Parse(strings.NewReader(`["FOO","BAR"]`))
				Expect(err).To(BeNil())
				Expect(got).To(Equal([]any{"FOO", "BAR"}))
			})
		})

		When("input is nested", func() {
			It("return expected object", func() {
				got, err := p.Parse(strings.NewReader(`[["FOO","BAR"],["BAZ","QUX"]]`))
				Expect(err).To(BeNil())
				Expect(got).To(Equal([]any{[]any{"FOO", "BAR"}, []any{"BAZ", "QUX"}}))
			})
		})

		When("input is mixed", func() {
			It("return expected object", func() {
				got, err := p.Parse(strings.NewReader(`["FOO",{"BAZ":"QUX"}]`))
				Expect(err).To(BeNil())
				Expect(got).To(Equal([]any{"FOO", map[string]any{"BAZ": "QUX"}}))
			})
		})

		When("input is multi_object", func() {
			It("return expected object", func() {
				got, err := p.Parse(strings.NewReader(`[{"FOO":"BAR"},{"BAZ":"QUX"}]`))
				Expect(err).To(BeNil())
				Expect(got).To(Equal([]any{map[string]any{"FOO": "BAR"}, map[string]any{"BAZ": "QUX"}}))
			})
		})
	})
})
