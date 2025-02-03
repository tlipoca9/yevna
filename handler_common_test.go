package yevna_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/tlipoca9/yevna"
	"github.com/tlipoca9/yevna/parser"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Common Handlers", func() {
	var buf *bytes.Buffer
	y := yevna.New()

	BeforeEach(func() {
		buf = &bytes.Buffer{}
	})

	Context("Handler - Chdir", func() {
		It("should change working directory for the command", func(ctx context.Context) {
			wd, err := os.Getwd()
			Expect(err).To(BeNil())

			// Check if windows
			switch runtime.GOOS {
			case "windows":
				err = y.Run(
					ctx,
					yevna.Chdir(filepath.Join(wd, "parser")),
					yevna.Exec("powershell", "-c", "Get-Location | Select-Object -ExpandProperty Path"),
					yevna.Tee(buf),
				)
				Expect(err).To(BeNil())
				Expect(buf.String()).To(Equal(filepath.Clean(filepath.Join(wd, "parser")) + "\r\n"))
			default:
				err = y.Run(
					ctx,
					yevna.Chdir(filepath.Join(wd, "parser")),
					yevna.Exec("pwd"),
					yevna.Tee(buf),
				)
				Expect(err).To(BeNil())
				Expect(buf.String()).To(Equal(filepath.Clean(filepath.Join(wd, "parser")) + "\n"))
			}
		})
	})

	Context("Handler - Silent", func() {
		It("should set the silent mode", func(ctx context.Context) {
			err := y.Run(
				ctx,
				yevna.Silent(true),
				yevna.Exec("curl", svc.URL+"/ipinfo"),
				yevna.Gjson("ip"),
				yevna.Tee(buf),
			)
			Expect(err).To(BeNil())
			Expect(buf.String()).To(Equal(`"1.1.1.1"`))
		})
	})

	Context("Handler - Input", func() {
		It("should success", func(ctx context.Context) {
			err := y.Run(
				ctx,
				yevna.Input(strings.NewReader("hello")),
				yevna.Exec("cat"),
				yevna.Tee(buf),
			)
			Expect(err).To(BeNil())
			Expect(buf.String()).To(Equal(`hello`))
		})
	})

	Context("Handler - Output", func() {
		It("should success", func(ctx context.Context) {
			var actual string
			err := y.Run(
				ctx,
				yevna.Input(strings.NewReader("hello")),
				yevna.Exec("cat"),
				yevna.ToStr(),
				yevna.Output(&actual),
			)
			Expect(err).To(BeNil())
			Expect(actual).To(Equal(`hello`))
		})
	})

	Context("Handler - Unmarshal", func() {
		When("input is json", func() {
			It("should return expected object", func(ctx context.Context) {
				var got map[string]any
				err := y.Run(
					ctx,
					yevna.Silent(true),
					yevna.Exec("curl", svc.URL+"/ipinfo"),
					yevna.Unmarshal(parser.JSON(), &got),
				)
				Expect(err).To(BeNil())
				Expect(got).To(Equal(ipInfoMap))
			})
		})
	})

	Context("Handler - OpenFile", func() {
		It("should success", func(ctx context.Context) {
			var got map[string]any
			err := y.Run(
				ctx,
				yevna.OpenFile("tests/test.json"),
				yevna.Unmarshal(parser.JSON(), &got),
			)
			Expect(err).To(BeNil())
			Expect(got).To(Equal(map[string]any{
				"name":  "Alice",
				"value": 42.,
			}))
		})
	})

	Context("Handler - WriteFile", func() {
		It("should success", func(ctx context.Context) {
			expectedContent := []byte(`{"name":"Alice","value":42}`)
			err := y.Run(
				ctx,
				yevna.Input(expectedContent),
				yevna.WriteFile("tests/write_tmp.json"),
			)
			Expect(err).To(BeNil())
			// Check if the file was written correctly
			fileContent, err := os.ReadFile("tests/write_tmp.json")
			Expect(err).To(BeNil())
			Expect(fileContent).To(Equal(expectedContent))
		})
	})

})
