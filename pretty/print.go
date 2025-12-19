package pretty

import (
	"bytes"
	"fmt"
	"io"
)

var _ io.Writer = &Printer{}

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
	msg := "\b\b\bOK \n%s"
	if userErr != nil {
		msg = "\b\b\bERROR\n%s"
	}
	if _, err := fmt.Fprintf(printer.stdout, msg, printer.buffer.String()); err != nil {
		return err
	}
	return nil
}
