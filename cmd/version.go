package cmd

import (
	"fmt"
	"io"
	"runtime/debug"
)

func Version() string {
	if info, ok := debug.ReadBuildInfo(); ok {
		return info.Main.Version
	}
	return "(unknown)"
}

type VersionCMD struct {
}

func (config *VersionCMD) Run(stdout io.Writer) error {
	_, err := fmt.Fprintln(stdout, Version())
	return err
}
