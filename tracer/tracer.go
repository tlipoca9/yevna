package tracer

import (
	"bytes"
	"fmt"
	"github.com/pterm/pterm"
	"io"
	"strings"
)

type Tracer interface {
	Trace(string)
}

type ExecTracer struct {
	secrets      map[string]string
	enableColor  bool
	w            io.Writer
	colorPrinter *pterm.BasicTextPrinter
}

type ExecTracerOption func(*ExecTracer)

func WithSecret(secretAndMasks ...string) ExecTracerOption {
	secrets := make(map[string]string)
	for i := 0; i < len(secretAndMasks); i += 2 {
		secrets[secretAndMasks[i]] = secretAndMasks[i+1]
	}
	return func(t *ExecTracer) {
		t.secrets = secrets
	}
}

func WithColor(enable bool) ExecTracerOption {
	return func(t *ExecTracer) {
		t.enableColor = enable
	}
}

func NewExecTracer(w io.Writer, opts ...ExecTracerOption) *ExecTracer {
	t := &ExecTracer{
		secrets:      make(map[string]string),
		enableColor:  true,
		w:            w,
		colorPrinter: &pterm.BasicTextPrinter{Writer: w},
	}
	for _, opt := range opts {
		opt(t)
	}
	return t
}

func (s *ExecTracer) Trace(cmd string) {
	// cmd may contain multiple commands separated by "|"
	commands := strings.Split(cmd, "|")
	var buf bytes.Buffer
	for i, command := range commands {
		if i == 0 {
			buf.WriteByte('$')
		} else {
			buf.WriteString(" |")
		}
		command = strings.TrimSpace(command)
		tokens := strings.Split(command, " ")
		for j, token := range tokens {
			buf.WriteByte(' ')
			token = strings.TrimSpace(token)

			// check if token is a secret
			if mask, ok := s.secrets[token]; ok {
				if s.enableColor {
					mask = pterm.FgLightRed.Sprint(mask)
				}
				buf.WriteString(mask)
				continue
			}

			// check if token is a multiline string
			if strings.Contains(token, "\n") {
				if strings.Contains(token, `"`) && strings.Contains(token, `'`) {
					token = fmt.Sprintf(`$'%s'`, token)
				} else if strings.Contains(token, `"`) {
					token = fmt.Sprintf(`'%s'`, token)
				} else {
					token = fmt.Sprintf(`"%s"`, token)
				}
			}

			// check if color is enabled
			if s.enableColor {
				if j == 0 {
					token = pterm.FgLightMagenta.Sprint(token)
				} else if token == "--" {
					token = pterm.FgLightCyan.Sprint(token)
				} else if strings.HasPrefix(token, "-") && !strings.HasPrefix(token, "---") {
					token = pterm.FgLightYellow.Sprint(token)
				}
			}
			buf.WriteString(token)
		}
	}
	buf.WriteByte('\n')

	if s.enableColor {
		s.colorPrinter.Print(buf.String())
	} else {
		_, _ = s.w.Write(buf.Bytes())
	}
}
