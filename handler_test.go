package yevna_test

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/tlipoca9/yevna"
	"github.com/tlipoca9/yevna/parser"
	"github.com/tlipoca9/yevna/tracer"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var (
	svc       *httptest.Server
	ipInfoMap = map[string]any{
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
	}
)

var _ = BeforeSuite(func() {
	gin.SetMode(gin.ReleaseMode)
	g := gin.New()
	g.GET("/ipinfo", func(c *gin.Context) {
		c.JSON(http.StatusOK, ipInfoMap)
	})
	svc = httptest.NewServer(g)
})

var _ = Describe("Handler", func() {
	var buf *bytes.Buffer
	y := yevna.New()
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

	Context("unmarshal", func() {
		When("input is json", func() {
			It("should return expected object", func() {
				var got map[string]any
				err := y.Run(
					context.Background(),
					yevna.Silent(true),
					yevna.Exec("curl", svc.URL+"/ipinfo"),
					yevna.Unmarshal(parser.JSON(), &got),
				)
				Expect(err).To(BeNil())
				Expect(got).To(Equal(ipInfoMap))
			})
		})
	})
})
