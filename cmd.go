package yevna

import (
	"bytes"
	"github.com/cockroachdb/errors"
	"github.com/go-viper/mapstructure/v2"
	"github.com/tlipoca9/yevna/parser"
	"github.com/tlipoca9/yevna/tracer"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Cmd enhances the exec.Cmd with more features
type Cmd struct {
	// Context is the context of the command
	*Context
	// cmd is the exec.Cmd will be executed
	cmd *exec.Cmd
	// prevProcess is the previous process in the pipeline
	// if it is nil, it means it is the first process of the pipeline
	prevProcess *Cmd
	// Err is the error of the command
	Err error
}

// Silent sets the silent mode
func (c *Cmd) Silent() *Cmd {
	if c.Err != nil {
		return c
	}

	for i := c.prevProcess; i != nil; i = i.prevProcess {
		i.cmd.Stderr = io.Discard
	}

	c.cmd.Stderr = io.Discard
	c.cmd.Stdout = io.Discard
	return c
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
	if c.cmd.Stdout != nil {
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
	c.Context.workDir = dir
	return c
}

// WithExecTrace sets the exec trace for the command
func (c *Cmd) WithExecTrace(enable bool) *Cmd {
	if c.Err != nil {
		return c
	}
	c.Context.enableExecTrace = enable
	return c
}

// WithExecTracer sets the exec tracer for the command
func (c *Cmd) WithExecTracer(t tracer.Tracer) *Cmd {
	if c.Err != nil {
		return c
	}
	c.Context.execTracer = t
	return c
}

// Run runs the command
// It will Start all the processes in the pipeline
func (c *Cmd) Run() error {
	if c.Err != nil {
		return c.Err
	}

	return c.Start().Wait().Err
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
	if c.Start().Err != nil {
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

	return c.Wait().Err
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
	return c.Start().Wait().Err
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
	return c.Start().Wait().Err
}

// Start starts the command
// It will Start all the processes in the pipeline
func (c *Cmd) Start() *Cmd {
	if c.Err != nil {
		return c
	}

	processes := make([]*Cmd, 0)
	for i := c.prevProcess; i != nil; i = i.prevProcess {
		processes = append(processes, i)
	}

	var err error
	for i := len(processes) - 1; i >= 0; i-- {
		if c.Context.workDir != "" {
			processes[i].cmd.Dir = c.Context.workDir
		}
		err = processes[i].cmd.Start()
		if err != nil {
			break
		}
	}
	if err != nil {
		for i := 0; i < len(processes); i++ {
			processes[i].Err = err
		}
		c.Err = err
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

	if c.Context.workDir != "" {
		c.cmd.Dir = c.Context.workDir
	}
	if c.Context.enableExecTrace {
		c.Context.execTracer.Trace(c.String())
	}
	c.Err = c.cmd.Start()
	return c
}

// Wait waits for the command
// It will Wait for all the processes in the pipeline
func (c *Cmd) Wait() *Cmd {
	if c.Err != nil {
		return c
	}

	if c.prevProcess != nil {
		err := c.prevProcess.Wait().Err
		if err != nil {
			c.Err = err
			return c
		}
	}

	c.Err = c.cmd.Wait()
	return c
}

// Pipe returns a new Cmd with the pipeline
func (c *Cmd) Pipe(name string, args ...string) *Cmd {
	if c.Err != nil {
		return c
	}

	if c.cmd.Stdout != nil {
		c.Err = ErrStdoutAlreadySet
		return c
	}

	stdoutPipe, err := c.cmd.StdoutPipe()
	if err != nil {
		c.Err = err
		return c
	}

	nextC := c.Context.Command(name, args...)
	nextC.prevProcess = c
	nextC.cmd.Stdin = stdoutPipe
	return nextC
}

func (c *Cmd) String() string {
	var buf bytes.Buffer
	if c.prevProcess != nil {
		buf.WriteString(c.prevProcess.String())
		buf.WriteString(" | ")
	}
	buf.WriteString(strings.Join(c.cmd.Args, " "))
	return buf.String()
}
