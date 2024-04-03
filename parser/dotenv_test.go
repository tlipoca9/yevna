package parser_test

import (
	"github.com/tlipoca9/yevna/parser"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("DotenvParser", func() {
	var got map[string]string

	BeforeEach(func() {
		got = nil
	})

	When("input is empty", func() {
		It("return empty object", func() {
			p := parser.Dotenv()
			err := p.Unmarshal([]byte(""), &got)
			Expect(err).To(BeNil())
			Expect(got).To(Equal(map[string]string{}))
		})
	})

	When("input is simple", func() {
		It("return expected object", func() {
			p := parser.Dotenv()
			err := p.Unmarshal([]byte("FOO=BAR\n"), &got)
			Expect(err).To(BeNil())
			Expect(got).To(Equal(map[string]string{"FOO": "BAR"}))
		})
	})

	When("input is multiline", func() {
		It("return expected object", func() {
			p := parser.Dotenv()
			err := p.Unmarshal([]byte("FOO=BAR\nBAZ=QUX\n"), &got)
			Expect(err).To(BeNil())
			Expect(got).To(Equal(map[string]string{"FOO": "BAR", "BAZ": "QUX"}))
		})
	})
})
