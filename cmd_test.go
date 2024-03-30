package yevna_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-viper/mapstructure/v2"

	"github.com/tlipoca9/yevna"
	"github.com/tlipoca9/yevna/parser"
	"github.com/tlipoca9/yevna/tracer"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Cmd", func() {
	var (
		buf *bytes.Buffer
		y   *yevna.Context
	)
	BeforeEach(func() {
		buf = &bytes.Buffer{}
		y = yevna.NewContext(
			context.Background(),
			yevna.WithExecTracer(tracer.NewExecTracer(buf, tracer.WithColor(false))),
		)
	})

	Context("Silent", func() {
		It("should set the silent mode", func() {
			err := y.Command("echo", "hello").Silent().Run()
			Expect(err).To(BeNil())
			Expect(buf.String()).To(Equal("$ echo hello\n"))
		})
	})

	Context("WithStdin", func() {
		It("should set the stdin for the command", func() {
			err := y.Command("cat").WithStdin(strings.NewReader("hello")).WithStdout(buf).Run()
			Expect(err).To(BeNil())
			Expect(buf.String()).To(Equal(`
$ cat
hello`[1:]))
		})
	})

	Context("WithStdout", func() {
		It("should set the stdout for the command", func() {
			err := y.Command("echo", "hello").WithStdout(buf).Run()
			Expect(err).To(BeNil())
			Expect(buf.String()).To(Equal(`
$ echo hello
hello
`[1:]))
		})
	})

	Context("WithWorkDir", func() {
		It("should set the working directory for the command", func() {
			wd, err := os.Getwd()
			Expect(err).To(BeNil())
			err = y.Command("pwd").WithWorkDir(filepath.Join(wd, "parser")).WithStdout(buf).Run()
			Expect(err).To(BeNil())
			Expect(buf.String()).To(Equal("$ pwd\n" + filepath.Join(wd, "parser") + "\n"))
		})
	})

	Context("WithExecTrace", func() {
		It("should disable exec trace for the command", func() {
			err := y.Command("echo", "hello").WithExecTrace(false).WithStdout(buf).Run()
			Expect(err).To(BeNil())
			Expect(buf.String()).To(Equal("hello\n"))
		})
	})

	Context("WithExecTracer", func() {
		It("should set the exec tracer for the command", func() {
			err := y.Command("echo", "hello").
				WithExecTracer(tracer.NewExecTracer(
					buf,
					tracer.WithColor(false),
					tracer.WithSecret("hello", "<world>"),
				)).
				WithStdout(buf).
				Run()
			Expect(err).To(BeNil())
			Expect(buf.String()).To(Equal("$ echo <world>\nhello\n"))
		})
	})

	Context("Run", func() {
		It("should run the command", func() {
			err := y.Command("echo", "hello").WithStdout(buf).Run()
			Expect(err).To(BeNil())
			Expect(buf.String()).To(Equal(`
$ echo hello
hello
`[1:]))
		})
	})

	Context("RunWithParser", func() {
		It("should run the command with the parser", func() {
			var res map[string]string
			err := y.Command("echo", `{"foo": "bar"}`).
				RunWithParser(parser.JSON(), &mapstructure.DecoderConfig{Result: &res})
			Expect(err).To(BeNil())
			Expect(buf.String()).To(Equal(`
$ echo {"foo": "bar"}
`[1:]))
			Expect(res).To(Equal(map[string]string{"foo": "bar"}))
		})
	})

	Context("Pipe", func() {
		It("should pipe the commands", func() {
			err := y.Command("echo", "hello,world,everyone").
				Pipe("xargs", "-d", ",", "-I", "{}", "echo", "{}").
				WithStdout(buf).
				Run()
			Expect(err).To(BeNil())
			Expect(buf.String()).To(Equal(`
$ echo hello,world,everyone | xargs -d , -I {} echo {}
hello
world
everyone

`[1:]))
		})
	})
})
