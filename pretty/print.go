package pretty

import (
	"bytes"
	"fmt"
	"io"

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
	_, err := fmt.Fprintf(printer.stdout, "Running %s: ...", printer.name)
	return err
}

func (printer *Printer) Done(userErr error) error {
	msg := "\b\b\bOK "
	print := OK
	if userErr != nil {
		msg = "\b\b\bERROR"
		print = ERROR
	}
	if _, err := print.Fprint(printer.stdout, msg); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(printer.stdout, printer.buffer.String()); err != nil {
		return err
	}
	return nil
}
