package yevna_test

import (
	"context"

	"github.com/tlipoca9/yevna"
	"github.com/tlipoca9/yevna/parser"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Handler - HTTP", func() {
	y := yevna.New()

	It("should success", func() {
		var got []string
		err := y.Run(
			context.Background(),
			yevna.Input(`[{"name": "Alice"}, {"name": "Bob"}]`),
			yevna.Gjson("#.name"),
			yevna.Unmarshal(parser.JSON(), &got),
		)
		Expect(err).To(BeNil())
		Expect(got).To(Equal([]string{"Alice", "Bob"}))
	})
})
