package watch

import (
	"bytes"
	"context"
	"fmt"
	"github.com/ohhfishal/gopher/report"
	"log/slog"
	"os/exec"
)

type BuildFunc func(context.Context, string) ([]report.BuildEvent, error)

func (config CMD) Build(ctx context.Context, logger *slog.Logger, build BuildFunc) {
	events, err := build(ctx, config.Path)
	if err != nil {
		logger.Error("go build", "err", err)
		return
	}
	logger.Info("built", "events", events)

	// Parse the output from JSON
	// Print pretty
	// Repeat with other steps
}

func goBuild(ctx context.Context, path string) ([]report.BuildEvent, error) {
	// TODO: Provide options to add more flags
	cmd := exec.CommandContext(ctx, "go", "build", "-json")
	cmd.Dir = path

	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() != 1 {
			return nil, fmt.Errorf("running command: %w: %s", err, string(output))
		}
	}
	if err == nil || len(output) == 0 {
		return []report.BuildEvent{}, nil
	}

	return report.ParseBuildJSON(bytes.NewReader(output))
}
