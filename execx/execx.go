package execx

import (
	"bytes"
	"context"
	"errors"
	"github.com/go-viper/mapstructure/v2"
	"github.com/pterm/pterm"
	"github.com/tlipoca9/yevna/parser"
	"os"
	"os/exec"
	"strings"
)

var GlobalOptions Options

type Options struct {
	quiet  func() bool
	print  func(a ...any)
	cmdStr func(cmd *exec.Cmd) string
}

func (o *Options) Quiet() {
	o.quiet = func() bool {
		return true
	}
}

func (o *Options) Verbose() {
	o.quiet = func() bool {
		return false
	}
}

func (o *Options) lazyInit() {
	if o.quiet == nil {
		o.quiet = GlobalOptions.quiet
	}
	if o.print == nil {
		o.print = GlobalOptions.print
	}
	if o.cmdStr == nil {
		o.cmdStr = GlobalOptions.cmdStr
	}
}

func init() {
	basicTxt := pterm.BasicTextPrinter{
		Writer: os.Stderr,
	}
	GlobalOptions.Verbose()
	GlobalOptions.print = func(a ...any) { basicTxt.Print(a...) }
	GlobalOptions.cmdStr = func(cmd *exec.Cmd) string {
		var buf bytes.Buffer
		name := cmd.Args[0]
		options := cmd.Args[1:]
		buf.WriteString("$ ")
		buf.WriteString(pterm.FgLightMagenta.Sprint(name))
		for _, v := range options {
			buf.WriteString(" ")
			if v == "--" {
				buf.WriteString(pterm.FgLightCyan.Sprint(v))
			} else if strings.HasPrefix(v, "-") && !strings.HasPrefix(v, "---") {
				buf.WriteString(pterm.FgLightYellow.Sprint(v))
			} else {
				buf.WriteString(v)
			}
		}
		buf.WriteString("\n")
		return buf.String()
	}
}

type Cmd struct {
	Options
	cmd *exec.Cmd

	parser        parser.Parser
	decoderConfig *mapstructure.DecoderConfig
}

func Command(ctx context.Context, name string, args ...string) *Cmd {
	if len(name) == 0 {
		panic("name is required, but got empty string")
	}
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return &Cmd{cmd: cmd}
}

func (c *Cmd) Quiet() *Cmd {
	c.Options.Quiet()
	return c
}

func (c *Cmd) printCmd() {
	c.lazyInit()
	if c.Options.quiet() {
		return
	}
	c.Options.print(c.Options.cmdStr(c.cmd))
}

func (c *Cmd) WithParser(p parser.Parser) *Cmd {
	c.parser = p
	return c
}

func (c *Cmd) WithDecoderConfig(dc mapstructure.DecoderConfig) *Cmd {
	c.decoderConfig = &dc
	return c
}

func (c *Cmd) Run() error {
	c.lazyInit()
	c.printCmd()

	if c.parser != nil {
		if c.decoderConfig == nil {
			return errors.New("DecodeConfig is required when Parser is set")
		}
		c.cmd.Stdout = nil
		rd, err := c.cmd.StdoutPipe()
		if err != nil {
			return err
		}
		defer rd.Close()
		err = c.cmd.Start()
		if err != nil {
			return err
		}
		data, err := c.parser.Parse(rd)
		if err != nil {
			return err
		}
		dec, err := mapstructure.NewDecoder(c.decoderConfig)
		if err != nil {
			return err
		}
		return dec.Decode(data)
	}

	return c.cmd.Run()
}
