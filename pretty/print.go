package pretty

import (
	"fmt"
	"io"
)

type Printer struct {
	name   string
	stdout io.Writer
}

func New(stdout io.Writer, name string) *Printer {
	printer := Printer{
		name:   name,
		stdout: stdout,
	}
	return &printer
}

func (printer *Printer) Start() error {
	_, err := fmt.Fprintf(printer.stdout, "Running %s: ...", printer.name)
	return err
}

func (printer *Printer) Done(err error) error {
	if err != nil {
		_, err := fmt.Fprintln(printer.stdout, "\b\b\bERROR\n"+err.Error())
		return err
	} else {
		_, err := fmt.Fprintln(printer.stdout, "\b\b\bOK ")
		return err
	}
}
