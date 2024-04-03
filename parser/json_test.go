package parser_test

import (
	"github.com/tlipoca9/yevna/parser"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("JSONParser", func() {
	Context("Object", func() {
		p := parser.JSON()
		var got map[string]any

		BeforeEach(func() {
			got = nil
		})

		When("input is empty", func() {
			It("return empty object", func() {
				buf := []byte("{}")
				err := p.Unmarshal(buf, &got)
				Expect(err).To(BeNil())
				Expect(got).To(BeEmpty())
			})
		})

		When("input is simple", func() {
			It("return expected object", func() {
				buf := []byte(`{"FOO":"BAR"}`)
				err := p.Unmarshal(buf, &got)
				Expect(err).To(BeNil())
				Expect(got).To(Equal(map[string]any{"FOO": "BAR"}))
			})
		})

		When("input is multiline", func() {
			It("return expected object", func() {
				buf := []byte(`{"FOO":"BAR","BAZ":"QUX"}`)
				err := p.Unmarshal(buf, &got)
				Expect(err).To(BeNil())
				Expect(got).To(Equal(map[string]any{"FOO": "BAR", "BAZ": "QUX"}))
			})
		})
	})

	Context("Array", func() {
		p := parser.JSON()
		var got []any

		BeforeEach(func() {
			got = nil
		})

		When("input is empty", func() {
			It("return empty object", func() {
				buf := []byte("[]")
				err := p.Unmarshal(buf, &got)
				Expect(err).To(BeNil())
				Expect(got).To(BeEmpty())
			})
		})

		When("input is simple", func() {
			It("return expected object", func() {
				buf := []byte(`["FOO"]`)
				err := p.Unmarshal(buf, &got)
				Expect(err).To(BeNil())
				Expect(got).To(Equal([]any{"FOO"}))
			})
		})

		When("input is multiline", func() {
			It("return expected object", func() {
				buf := []byte(`["FOO","BAR"]`)
				err := p.Unmarshal(buf, &got)
				Expect(err).To(BeNil())
				Expect(got).To(Equal([]any{"FOO", "BAR"}))
			})
		})

		When("input is nested", func() {
			It("return expected object", func() {
				buf := []byte(`[["FOO","BAR"],["BAZ","QUX"]]`)
				err := p.Unmarshal(buf, &got)
				Expect(err).To(BeNil())
				Expect(got).To(Equal([]any{[]any{"FOO", "BAR"}, []any{"BAZ", "QUX"}}))
			})
		})

		When("input is mixed", func() {
			It("return expected object", func() {
				buf := []byte(`["FOO",{"BAZ":"QUX"}]`)
				err := p.Unmarshal(buf, &got)
				Expect(err).To(BeNil())
				Expect(got).To(Equal([]any{"FOO", map[string]any{"BAZ": "QUX"}}))
			})
		})

		When("input is multi_object", func() {
			It("return expected object", func() {
				buf := []byte(`[{"FOO":"BAR"},{"BAZ":"QUX"}]`)
				err := p.Unmarshal(buf, &got)
				Expect(err).To(BeNil())
				Expect(got).To(Equal([]any{map[string]any{"FOO": "BAR"}, map[string]any{"BAZ": "QUX"}}))
			})
		})
	})
})
