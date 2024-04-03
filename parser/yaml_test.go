package parser_test

import (
	"github.com/tlipoca9/yevna/parser"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("YAMLParser", func() {
	Context("Object", func() {
		p := parser.YAML()
		var got map[string]any

		BeforeEach(func() {
			got = nil
		})

		When("input is empty", func() {
			It("return empty object", func() {
				buf := []byte("")

				err := p.Unmarshal(buf, &got)
				Expect(err).To(BeNil())
				Expect(got).To(BeEmpty())
			})
		})

		When("input is simple", func() {
			It("return expected object", func() {
				buf := []byte("FOO: BAR\n")

				err := p.Unmarshal(buf, &got)
				Expect(err).To(BeNil())
				Expect(got).To(Equal(map[string]any{"FOO": "BAR"}))
			})
		})

		When("input is multiline", func() {
			It("return expected object", func() {
				buf := []byte("FOO: BAR\nBAZ: QUX\n")

				err := p.Unmarshal(buf, &got)
				Expect(err).To(BeNil())
				Expect(got).To(Equal(map[string]any{"FOO": "BAR", "BAZ": "QUX"}))
			})
		})
	})

	Context("Array", func() {
		p := parser.YAML()
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
				buf := []byte("- FOO\n")

				err := p.Unmarshal(buf, &got)
				Expect(err).To(BeNil())
				Expect(got).To(Equal([]any{"FOO"}))
			})
		})

		When("input is multiline", func() {
			It("return expected object", func() {
				buf := []byte("- FOO\n- BAR\n")

				err := p.Unmarshal(buf, &got)
				Expect(err).To(BeNil())
				Expect(got).To(Equal([]any{"FOO", "BAR"}))
			})
		})
	})
})
