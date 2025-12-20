package cache

import (
	"fmt"
	"io"
	"log/slog"
	"os/exec"
	"runtime/debug"
	"strings"
	"time"
)

func Version() string {
	if info, ok := debug.ReadBuildInfo(); ok {
		return info.Main.Version
	}
	return "(unknown)"
}

type CMD struct {
}

func (config *CMD) Run(stdout io.Writer) error {
	_, err := fmt.Fprintln(stdout, Version())
	return err
}

func GoVersion(goBin string) string {
	cmd := exec.Command(goBin, "version")

	output, err := cmd.Output()
	if err != nil {
		slog.Warn("could not get go version", "err", err)
		return fmt.Sprintf("unknown: %d", time.Now().Unix())
	}

	if version := string(output); version != "" {
		return strings.TrimSpace(version)
	}
	return fmt.Sprintf("unknown: %d", time.Now().Unix())
}
