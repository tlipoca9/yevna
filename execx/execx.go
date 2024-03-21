package execx

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"strings"

	"github.com/go-viper/mapstructure/v2"
	"github.com/pterm/pterm"

	"github.com/tlipoca9/yevna/parser"
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
	cmd         *exec.Cmd
	pipeProcess *Cmd

	Err error
	Options
	parser        parser.Parser
	decoderConfig *mapstructure.DecoderConfig
}

func Command(ctx context.Context, name string, args ...string) *Cmd {
	if len(name) == 0 {
		return &Cmd{Err: errors.New("name is required, but got empty string")}
	}
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return &Cmd{
		cmd: cmd,
		Options: Options{
			quiet:  GlobalOptions.quiet,
			print:  GlobalOptions.print,
			cmdStr: GlobalOptions.cmdStr,
		},
	}
}

func Pipe(ctx context.Context, cmds ...[]string) *Cmd {
	if len(cmds) == 0 {
		return &Cmd{Err: errors.New("at least one command is required")}
	}

	c := Command(ctx, cmds[0][0], cmds[0][1:]...)
	for i := 1; i < len(cmds); i++ {
		c = c.Pipe(ctx, cmds[i][0], cmds[i][1:]...)
	}

	return c
}

func (c *Cmd) Quiet() *Cmd {
	if c.Err != nil {
		return c
	}

	c.Options.Quiet()
	return c
}

func (c *Cmd) printCmd() {
	if c.Err != nil {
		return
	}

	if c.Options.quiet() {
		return
	}
	c.Options.print(c.Options.cmdStr(c.cmd))
}

func (c *Cmd) WithParser(p parser.Parser) *Cmd {
	if c.Err != nil {
		return c
	}

	c.parser = p
	return c
}

func (c *Cmd) WithDecoderConfig(dc mapstructure.DecoderConfig) *Cmd {
	if c.Err != nil {
		return c
	}

	c.decoderConfig = &dc
	return c
}

func (c *Cmd) Run() *Cmd {
	if c.Err != nil {
		return c
	}

	if c.parser != nil {
		if c.decoderConfig == nil {
			c.Err = errors.New("DecodeConfig is required when Parser is set")
			return c
		}
		c.cmd.Stdout = nil
		rd, err := c.cmd.StdoutPipe()
		if err != nil {
			c.Err = err
			return c
		}
		defer rd.Close()
		if c.start().Err != nil {
			return c
		}
		data, err := c.parser.Parse(rd)
		if err != nil {
			c.Err = err
			return c
		}
		dec, err := mapstructure.NewDecoder(c.decoderConfig)
		if err != nil {
			c.Err = err
			return c
		}

		err = dec.Decode(data)
		if err != nil {
			c.Err = err
			return c
		}

		return c.wait()
	}

	return c.start().wait()
}

func (c *Cmd) start() *Cmd {
	if c.Err != nil {
		return c
	}

	// start all the processes in the pipeline
	var processes []*Cmd
	for i := c.pipeProcess; i != nil; i = i.pipeProcess {
		processes = append(processes, i)
	}

	var err error
	for i := len(processes) - 1; i >= 0; i-- {
		processes[i].printCmd()
		err = processes[i].cmd.Start()
		if err != nil {
			break
		}
	}
	for i := len(processes) - 1; i >= 0; i-- {
		processes[i].Err = err
	}
	if err != nil {
		c.Err = err
		return c
	}

	c.printCmd()
	c.Err = c.cmd.Start()
	return c
}

func (c *Cmd) wait() *Cmd {
	if c.Err != nil {
		return c
	}

	// wait for all the processes in the pipeline
	var processes []*Cmd
	for i := c.pipeProcess; i != nil; i = i.pipeProcess {
		processes = append(processes, i)
	}
	var err error
	for i := len(processes) - 1; i >= 0; i-- {
		err = processes[i].cmd.Wait()
		if err != nil {
			break
		}
	}
	for i := len(processes) - 1; i >= 0; i-- {
		processes[i].Err = err
	}
	if err != nil {
		c.Err = err
		return c
	}

	c.Err = c.cmd.Wait()
	return c
}

func (c *Cmd) Pipe(ctx context.Context, name string, args ...string) *Cmd {
	if c.Err != nil {
		return c
	}

	if c.parser != nil {
		return &Cmd{Err: errors.New("cannot pipe after setting a parser")}
	}
	if len(name) == 0 {
		return &Cmd{Err: errors.New("name is required, but got empty string")}
	}

	c.cmd.Stdout = nil

	prevStdout, err := c.cmd.StdoutPipe()
	if err != nil {
		c.Err = err
		return c
	}

	nextC := Command(ctx, name, args...)
	nextC.pipeProcess = c
	nextC.cmd.Stdin = prevStdout
	return nextC
}
