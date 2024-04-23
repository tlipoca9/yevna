package tracer

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/pterm/pterm"
)

type Tracer interface {
	Trace(name string, args ...string)
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
		if i+1 >= len(secretAndMasks) {
			secrets[secretAndMasks[i]] = "******"
			break
		}
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

func (s *ExecTracer) Trace(name string, args ...string) {
	var buf bytes.Buffer
	buf.WriteByte('$')
	buf.WriteByte(' ')
	if s.enableColor {
		buf.WriteString(pterm.FgLightMagenta.Sprint(name))
	} else {
		buf.WriteString(name)
	}
	for _, arg := range args {
		buf.WriteByte(' ')
		arg = strings.TrimSpace(arg)

		// check if arg is a secret
		if mask, ok := s.secrets[arg]; ok {
			if s.enableColor {
				mask = pterm.FgLightRed.Sprint(mask)
			}
			buf.WriteString(mask)
			continue
		}

		// check if arg is a multiline string
		if strings.ContainsAny(arg, " \n") {
			if strings.Contains(arg, `"`) && strings.Contains(arg, `'`) {
				arg = fmt.Sprintf(`$'%s'`, arg)
			} else if strings.Contains(arg, `"`) {
				arg = fmt.Sprintf(`'%s'`, arg)
			} else {
				arg = fmt.Sprintf(`"%s"`, arg)
			}
		}

		// check if color is enabled
		if s.enableColor {
			if arg == "--" {
				arg = pterm.FgLightCyan.Sprint(arg)
			} else if strings.HasPrefix(arg, "-") && !strings.HasPrefix(arg, "---") {
				arg = pterm.FgLightYellow.Sprint(arg)
			}
		}
		buf.WriteString(arg)
	}
	buf.WriteByte('\n')

	if s.enableColor {
		s.colorPrinter.Print(buf.String())
	} else {
		_, _ = s.w.Write(buf.Bytes())
	}
}
