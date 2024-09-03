package yevna_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"

	"github.com/tlipoca9/yevna"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Handler - Chdir", func() {
	var buf *bytes.Buffer
	y := yevna.New()

	BeforeEach(func() {
		buf = &bytes.Buffer{}
	})

	It("should change working directory for the command", func() {
		wd, err := os.Getwd()
		Expect(err).To(BeNil())
		err = y.Run(
			context.Background(),
			yevna.Chdir(filepath.Join(wd, "parser")),
			yevna.Exec("pwd"),
			yevna.Tee(buf),
		)
		Expect(err).To(BeNil())
		Expect(buf.String()).To(Equal(filepath.Clean(filepath.Join(wd, "parser")) + "\n"))
	})
})
