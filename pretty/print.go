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
var ERROR = color.New(color.FgRed)

type Printer struct {
	name   string
	stdout io.Writer
	buffer bytes.Buffer
}

func New(stdout io.Writer, name string) *Printer {
	printer := Printer{
		name:   name,
		stdout: stdout,
	}
	return &printer
}

func (printer *Printer) Write(content []byte) (int, error) {
	return printer.buffer.Write(content)
}

func (printer *Printer) Start() error {
	_, err := fmt.Fprintf(printer.stdout, "Running %s: ", printer.name)
	return err
}

func (printer *Printer) Done(userErr error) error {
	msg := OK.Sprint("OK")
	if userErr != nil {
		msg = ERROR.Sprint("ERROR")
	}
	if _, err := fmt.Fprintln(printer.stdout, msg); err != nil {
		return err
	}
	output := printer.buffer.String()
	for line := range strings.Lines(output) {
		line = RemoveTrailingSpaces(line)
		if len(line) == 0 {
			continue
		}
		if _, err := fmt.Fprintln(printer.stdout, line); err != nil {
			return err
		}
	}
	return nil
}

func RemoveTrailingSpaces(str string) string {
	return strings.TrimRightFunc(str, func(r rune) bool {
		return r == '\t' || r == '\n' || r == ' '
	})
}
