package yevna

import (
	"bytes"
	"context"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/go-viper/mapstructure/v2"
	"github.com/pterm/pterm"

	"github.com/tlipoca9/yevna/parser"
)

// GlobalOptions is the global options for the command
// It is used to set the default options for the command
var GlobalOptions Options

// Options is the options for the command
type Options struct {
	quiet  func() bool
	print  func(a ...any)
	cmdStr func(cmd *exec.Cmd) string
}

// Quiet sets the quiet mode
func (o *Options) Quiet() {
	o.quiet = func() bool {
		return true
	}
}

// Verbose sets the verbose mode
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

// Cmd enhances the exec.Cmd with more features
type Cmd struct {
	// cmd is the exec.Cmd will be executed
	cmd *exec.Cmd
	// prevProcess is the previous process in the pipeline
	// if it is nil, it means it is the first process of the pipeline
	prevProcess *Cmd
	// Err is the error of the command
	Err error
	// Options is the options for the command
	Options
}

// Command returns a new Cmd with the global options
func Command(ctx context.Context, name string, args ...string) *Cmd {
	if len(name) == 0 {
		return &Cmd{Err: ErrNameRequired}
	}
	cmd := exec.CommandContext(ctx, name, args...)
	return &Cmd{
		cmd: cmd,
		Options: Options{
			quiet:  GlobalOptions.quiet,
			print:  GlobalOptions.print,
			cmdStr: GlobalOptions.cmdStr,
		},
	}
}

// Pipe returns a new Cmd with the pipeline
// The first command is the first element of the cmds
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

// Quiet sets the quiet mode
func (c *Cmd) Quiet() *Cmd {
	if c.Err != nil {
		return c
	}
	if c.prevProcess != nil {
		c.prevProcess.Quiet()
	}

	c.Options.Quiet()
	return c
}

// printCmd prints the command before executing it
func (c *Cmd) printCmd() {
	if c.Err != nil {
		return
	}

	if c.Options.quiet() {
		return
	}
	c.Options.print(c.Options.cmdStr(c.cmd))
}

// WithStdin sets the stdin for the command
func (c *Cmd) WithStdin(r io.Reader) *Cmd {
	if c.Err != nil {
		return c
	}

	if c.prevProcess != nil {
		c.Err = errors.New("stdin is not allowed in pipeline")
		return c
	}

	c.cmd.Stdin = r
	return c
}

// WithStdout sets the stdout for the command
func (c *Cmd) WithStdout(w io.Writer) *Cmd {
	if c.Err != nil {
		return c
	}
	if c.cmd.Stdout != nil && c.cmd.Stdout != os.Stdout {
		c.Err = ErrStdoutAlreadySet
		return c
	}

	c.cmd.Stdout = w
	return c
}

// WithStderr sets the stderr for the command
func (c *Cmd) WithStderr(w io.Writer) *Cmd {
	if c.Err != nil {
		return c
	}
	if c.prevProcess != nil {
		c.prevProcess.WithStderr(w)
	}

	c.cmd.Stderr = w
	return c
}

// WithWorkDir sets the working directory for the command
func (c *Cmd) WithWorkDir(dir string) *Cmd {
	if c.Err != nil {
		return c
	}
	if c.prevProcess != nil {
		c.prevProcess.WithWorkDir(dir)
	}

	c.cmd.Dir = dir
	return c
}

// Run runs the command
// It will start all the processes in the pipeline
func (c *Cmd) Run() error {
	if c.Err != nil {
		return c.Err
	}

	return c.start().wait().Err
}

// RunWithParser runs the command with the parser
// It parses the stdout of the command with the parser
// and decodes the data with the decoder
func (c *Cmd) RunWithParser(p parser.Parser, dc *mapstructure.DecoderConfig) error {
	if c.Err != nil {
		return c.Err
	}

	if c.cmd.Stdout != nil {
		c.Err = ErrStdoutAlreadySet
		return c.Err
	}
	rd, err := c.cmd.StdoutPipe()
	if err != nil {
		c.Err = err
		return c.Err
	}
	defer rd.Close()
	if c.start().Err != nil {
		return c.Err
	}
	data, err := p.Parse(rd)
	if err != nil {
		c.Err = err
		return c.Err
	}
	dec, err := mapstructure.NewDecoder(dc)
	if err != nil {
		c.Err = err
		return c.Err
	}

	err = dec.Decode(data)
	if err != nil {
		c.Err = err
		return c.Err
	}

	return c.wait().Err
}

// RunWithParseFunc is shorthand for RunWithParser with the parse function
func (c *Cmd) RunWithParseFunc(p func(r io.Reader) (any, error), dc *mapstructure.DecoderConfig) error {
	if c.Err != nil {
		return c.Err
	}

	return c.RunWithParser(parser.ParseFunc(p), dc)
}

// WriteFile runs the command and writes the stdout to the file
func (c *Cmd) WriteFile(path string) error {
	if c.Err != nil {
		return c.Err
	}

	path, err := filepath.Abs(path)
	if err != nil {
		c.Err = err
		return c.Err
	}
	err = os.MkdirAll(filepath.Dir(path), os.FileMode(0755))
	if err != nil {
		c.Err = err
	}

	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(0644))
	if err != nil {
		c.Err = err
		return c.Err
	}
	defer f.Close()

	if c.cmd.Stdout != nil {
		c.Err = ErrStdoutAlreadySet
		return c.Err
	}
	c.cmd.Stdout = f
	return c.start().wait().Err
}

// AppendFile runs the command and appends the stdout to the file
func (c *Cmd) AppendFile(path string) error {
	if c.Err != nil {
		return c.Err
	}

	path, err := filepath.Abs(path)
	if err != nil {
		c.Err = err
		return c.Err
	}
	err = os.MkdirAll(filepath.Dir(path), os.FileMode(0755))
	if err != nil {
		c.Err = err
	}

	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, os.FileMode(0644))
	if err != nil {
		c.Err = err
		return c.Err
	}

	if c.cmd.Stdout != nil {
		c.Err = ErrStdoutAlreadySet
		return c.Err
	}
	c.cmd.Stdout = f
	return c.start().wait().Err
}

// start starts the command
// It will start all the processes in the pipeline
func (c *Cmd) start() *Cmd {
	if c.Err != nil {
		return c
	}
	if c.cmd.Stdin == nil {
		c.cmd.Stdin = os.Stdin
	}
	if c.cmd.Stdout == nil {
		c.cmd.Stdout = os.Stdout
	}
	if c.cmd.Stderr == nil {
		c.cmd.Stderr = os.Stderr
	}

	// start all the processes in the pipeline
	var processes []*Cmd
	for i := c.prevProcess; i != nil; i = i.prevProcess {
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

// wait waits for the command
// It will wait for all the processes in the pipeline
func (c *Cmd) wait() *Cmd {
	if c.Err != nil {
		return c
	}

	// wait for all the processes in the pipeline
	var processes []*Cmd
	for i := c.prevProcess; i != nil; i = i.prevProcess {
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

// Pipe returns a new Cmd with the pipeline
func (c *Cmd) Pipe(ctx context.Context, name string, args ...string) *Cmd {
	if c.Err != nil {
		return c
	}

	if len(name) == 0 {
		return &Cmd{Err: ErrNameRequired}
	}

	if c.cmd.Stdout != nil {
		c.Err = ErrStdoutAlreadySet
		return c
	}
	c.cmd.Stdout = nil

	prevStdout, err := c.cmd.StdoutPipe()
	if err != nil {
		c.Err = err
		return c
	}

	nextC := Command(ctx, name, args...)
	nextC.prevProcess = c
	nextC.cmd.Stdin = prevStdout
	return nextC
}
