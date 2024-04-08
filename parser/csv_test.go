package parser_test

import (
	"github.com/tlipoca9/yevna/parser"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("CsvParser", func() {
	var got []map[string]string

	BeforeEach(func() {
		got = nil
	})

	When("input is empty", func() {
		It("return empty object", func() {
			p := parser.CSV()
			err := p.Unmarshal([]byte(""), &got)
			Expect(err).To(BeNil())
			Expect(len(got)).To(Equal(0))
		})
	})

	When("input just has headers", func() {
		It("return empty object", func() {
			p := parser.CSV()
			err := p.Unmarshal([]byte(`
FOO,BAR
`[1:]), &got)
			Expect(err).To(BeNil())
			Expect(len(got)).To(Equal(0))
		})
	})

	When("input is simple", func() {
		It("return expected object", func() {
			p := parser.CSV()
			err := p.Unmarshal([]byte(`
FOO,BAR
42,4242
`[1:]), &got)
			Expect(err).To(BeNil())
			Expect(got).To(Equal([]map[string]string{
				{"FOO": "42", "BAR": "4242"},
			}))
		})
	})

	When("input is multiline", func() {
		It("return expected object", func() {
			p := parser.CSV()
			err := p.Unmarshal([]byte(`
FOO,BAR
42,4242
123,456
`[1:]), &got)
			Expect(err).To(BeNil())
			Expect(got).To(Equal([]map[string]string{
				{"FOO": "42", "BAR": "4242"},
				{"FOO": "123", "BAR": "456"},
			}))
		})
	})

	When("input is multiline with headers", func() {
		It("return expected object", func() {
			p := parser.CSV().WithHeaders("BAR", "FOO")
			err := p.Unmarshal([]byte(`
123,456
qwe,asd
`[1:]), &got)
			Expect(err).To(BeNil())
			Expect(got).To(Equal([]map[string]string{
				{"FOO": "456", "BAR": "123"},
				{"FOO": "asd", "BAR": "qwe"},
			}))
		})
	})
})
