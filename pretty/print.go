package pretty

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/fatih/color"
)

var _ io.Writer = &Printer{}

var OK = color.New(color.FgGreen)
var WARN = color.New(color.FgYellow)
var ERROR = color.New(color.FgRed)

var OkText = OK.Sprintf("OK")
var ErrorText = ERROR.Sprintf("ERROR")
var WarnText = WARN.Sprintf("WARN")

type Printer struct {
	name     string
	stdout   io.Writer
	buffer   bytes.Buffer
	indent   string
	warnings []error
}

func New(stdout io.Writer, name string, delim ...string) *Printer {
	printer := Printer{
		name:   name,
		stdout: stdout,
	}
	if len(delim) >= 1 {
		printer.indent = delim[0]
	}
	return &printer
}

func (printer *Printer) Write(content []byte) (int, error) {
	return printer.buffer.Write(content)
}

func (printer *Printer) Start() error {
	_, err := fmt.Fprintf(printer.stdout, "%s: ", printer.name)
	return err
}

func (printer *Printer) Done(userErr error) error {
	msg := OkText
	if userErr != nil {
		msg = ErrorText
	} else if len(printer.warnings) > 0 {
		msg = WARN.Sprintf("WARN (%d)", len(printer.warnings))
	}
	if _, err := fmt.Fprintln(printer.stdout, msg); err != nil {
		return err
	}

	stdout := NewIndentedWriter(printer.stdout, printer.indent)

	for i, warning := range printer.warnings {
		fmt.Fprintf(stdout, "%s %s\n", WARN.Sprintf("(%d)", i+1), warning.Error())
	}

	output := printer.buffer.String()
	for line := range strings.Lines(output) {
		line = RemoveTrailingSpaces(line)
		if len(line) == 0 {
			continue
		}
		if _, err := fmt.Fprintln(stdout, line); err != nil {
			return err
		}
	}
	return nil
}

func (printer *Printer) Warn(warnings ...error) {
	for _, warning := range warnings {
		printer.warnings = append(printer.warnings, warning)
	}
}

func RemoveTrailingSpaces(str string) string {
	return strings.TrimRightFunc(str, func(r rune) bool {
		return r == '\t' || r == '\n' || r == ' '
	})
}
