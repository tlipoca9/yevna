package yevna_test

import (
	"bytes"
	"context"
	"github.com/gin-gonic/gin"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"

	"github.com/tlipoca9/yevna"
	"github.com/tlipoca9/yevna/tracer"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var svc *httptest.Server

var _ = BeforeSuite(func() {
	gin.SetMode(gin.ReleaseMode)
	g := gin.New()
	g.GET("/ipinfo", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"ip":       "1.1.1.1",
			"hostname": "one.one.one.one",
			"anycast":  true,
			"city":     "The Rocks",
			"region":   "New South Wales",
			"country":  "AU",
			"loc":      "-33.8592,151.2081",
			"org":      "AS13335 Cloudflare, Inc.",
			"postal":   "2000",
			"timezone": "Australia/Sydney",
		})
	})
	svc = httptest.NewServer(g)
})

var _ = Describe("Handler", func() {
	var buf *bytes.Buffer
	y := yevna.NewHandlerContext()
	y.Tracer(tracer.Discard)

	BeforeEach(func() {
		buf = &bytes.Buffer{}
	})

	Context("Silent", func() {
		It("should set the silent mode", func() {
			err := y.Run(
				context.Background(),
				yevna.Silent(true),
				yevna.Exec("curl", svc.URL+"/ipinfo"),
				yevna.Exec("jq", ".ip"),
				yevna.Tee(buf),
			)
			Expect(err).To(BeNil())
			Expect(buf.String()).To(Equal(`"1.1.1.1"` + "\n"))
		})
	})

	Context("Echo", func() {
		It("should set the stdin for the command", func() {
			err := y.Run(
				context.Background(),
				yevna.Echo(strings.NewReader("hello")),
				yevna.Exec("cat"),
				yevna.Tee(buf),
			)
			Expect(err).To(BeNil())
			Expect(buf.String()).To(Equal(`hello`))
		})
	})

	Context("normal", func() {
		It("should set the stdout for the command", func() {
			err := y.Run(
				context.Background(),
				yevna.Exec("echo", "hello"),
				yevna.Tee(buf),
			)
			Expect(err).To(BeNil())
			Expect(buf.String()).To(Equal("hello\n"))
		})
	})

	Context("chdir", func() {
		It("should set the working directory for the command", func() {
			wd, err := os.Getwd()
			Expect(err).To(BeNil())
			err = y.Run(
				context.Background(),
				yevna.Cd(filepath.Join(wd, "parser")),
				yevna.Exec("pwd"),
				yevna.Tee(buf),
			)
			Expect(err).To(BeNil())
			Expect(buf.String()).To(Equal(filepath.Join(wd, "parser") + "\n"))
		})
	})

})
