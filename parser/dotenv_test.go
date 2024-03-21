package parser_test

import (
	"strings"

	"github.com/tlipoca9/yevna/parser"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("DotenvParser", func() {
	When("input is empty", func() {
		It("return empty object", func() {
			p := parser.Dotenv()
			got, err := p.Parse(strings.NewReader(""))
			Expect(err).To(BeNil())
			Expect(got).To(Equal(map[string]string{}))
		})
	})

	When("input is simple", func() {
		It("return expected object", func() {
			p := parser.Dotenv()
			got, err := p.Parse(strings.NewReader("FOO=BAR\n"))
			Expect(err).To(BeNil())
			Expect(got).To(Equal(map[string]string{"FOO": "BAR"}))
		})
	})

	When("input is multiline", func() {
		It("return expected object", func() {
			p := parser.Dotenv()
			got, err := p.Parse(strings.NewReader("FOO=BAR\nBAZ=QUX\n"))
			Expect(err).To(BeNil())
			Expect(got).To(Equal(map[string]string{"FOO": "BAR", "BAZ": "QUX"}))
		})
	})
})
