package yevna_test

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/goccy/go-json"
	"github.com/imroc/req/v3"

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
	g := http.NewServeMux()
	g.HandleFunc("/ipinfo", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		buf, err := json.Marshal(ipInfoMap)
		if err != nil {
			_, _ = w.Write([]byte(err.Error()))
			return
		}
		_, _ = w.Write(buf)
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

	Context("Cd", func() {
		It("should change working directory for the command", func() {
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

	Context("Silent", func() {
		It("should set the silent mode", func() {
			err := y.Run(
				context.Background(),
				yevna.Silent(true),
				yevna.Exec("curl", svc.URL+"/ipinfo"),
				yevna.Gjson("ip"),
				yevna.Tee(buf),
			)
			Expect(err).To(BeNil())
			Expect(buf.String()).To(Equal(`"1.1.1.1"`))
		})
	})

	Context("WithReader", func() {
		It("should success", func() {
			err := y.Run(
				context.Background(),
				yevna.WithReader(strings.NewReader("hello")),
				yevna.Exec("cat"),
				yevna.Tee(buf),
			)
			Expect(err).To(BeNil())
			Expect(buf.String()).To(Equal(`hello`))
		})
	})

	Context("Echo", func() {
		It("should success", func() {
			err := y.Run(
				context.Background(),
				yevna.Echo("hello"),
				yevna.Tee(buf),
			)
			Expect(err).To(BeNil())
			Expect(buf.String()).To(Equal("hello"))
		})
	})

	Context("Exec", func() {
		It("should success", func() {
			err := y.Run(
				context.Background(),
				yevna.Exec("echo", "hello"),
				yevna.Tee(buf),
			)
			Expect(err).To(BeNil())
			Expect(buf.String()).To(Equal("hello\n"))
		})
	})

	Context("Execs", func() {
		It("should success", func() {
			err := y.Run(
				context.Background(),
				yevna.Execs("echo 'hello world'"),
				yevna.Tee(buf),
			)
			Expect(err).To(BeNil())
			Expect(buf.String()).To(Equal("hello world\n"))
		})
	})

	Context("Execf", func() {
		It("should success", func() {
			err := y.Run(
				context.Background(),
				yevna.Execf("echo '%s %s'", "hello", "world"),
				yevna.Tee(buf),
			)
			Expect(err).To(BeNil())
			Expect(buf.String()).To(Equal("hello world\n"))
		})
	})

	Context("Unmarshal", func() {
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

	Context("HTTP", func() {
		It("should success", func() {
			var got map[string]any
			err := y.Run(
				context.Background(),
				yevna.HTTP().MakeRequest(func(c *req.Client, _ any) *req.Request {
					return c.SetTimeout(time.Second).Get(svc.URL + "/ipinfo")
				}),
				yevna.Unmarshal(parser.JSON(), &got),
			)
			Expect(err).To(BeNil())
			Expect(got).To(Equal(ipInfoMap))
		})
	})

	Context("Gjson", func() {
		It("should success", func() {
			var got []string
			err := y.Run(
				context.Background(),
				yevna.Echo(`[{"name": "Alice"}, {"name": "Bob"}]`),
				yevna.Gjson("#.name"),
				yevna.Unmarshal(parser.JSON(), &got),
			)
			Expect(err).To(BeNil())
			Expect(got).To(Equal([]string{"Alice", "Bob"}))
		})
	})
})
