package yevna_test

import (
	"context"
	"net/http"

	"github.com/tlipoca9/yevna"
	"github.com/tlipoca9/yevna/parser"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Handler - HTTP", func() {
	y := yevna.New()

	It("should success", func(ctx context.Context) {
		var got map[string]any
		err := y.Run(
			ctx,
			yevna.HTTP(func(c *yevna.Context, in any) (*http.Request, error) {
				return http.NewRequest(http.MethodGet, svc.URL+"/ipinfo", nil)
			}),
			yevna.Unmarshal(parser.JSON(), &got),
		)
		Expect(err).To(BeNil())
		Expect(got).To(Equal(ipInfoMap))
	})
})
