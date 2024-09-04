package yevna_test

import (
	"bytes"
	"context"

	"github.com/tlipoca9/yevna"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Handler - Exec", func() {
	var buf *bytes.Buffer
	y := yevna.New()

	BeforeEach(func() {
		buf = &bytes.Buffer{}
	})

	Context("Exec", func() {
		It("should success", func(ctx context.Context) {
			err := y.Run(
				ctx,
				yevna.Exec("echo", "hello"),
				yevna.Tee(buf),
			)
			Expect(err).To(BeNil())
			Expect(buf.String()).To(Equal("hello\n"))
		})
	})

	Context("Execs", func() {
		It("should success", func(ctx context.Context) {
			err := y.Run(
				ctx,
				yevna.Execs("echo 'hello world'"),
				yevna.Tee(buf),
			)
			Expect(err).To(BeNil())
			Expect(buf.String()).To(Equal("hello world\n"))
		})
	})

	Context("Execf", func() {
		It("should success", func(ctx context.Context) {
			err := y.Run(
				ctx,
				yevna.Execf("echo '%s %s'", "hello", "world"),
				yevna.Tee(buf),
			)
			Expect(err).To(BeNil())
			Expect(buf.String()).To(Equal("hello world\n"))
		})
	})
})
