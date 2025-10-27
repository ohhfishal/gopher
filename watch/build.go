package watch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os/exec"
)

type BuildFunc func(context.Context, string) ([]BuildEvent, error)

// From https://pkg.go.dev/cmd/go#hdr-Build__json_encoding
type BuildEvent struct {
	// TODO: Get the import path using go list -json. Then use that to truncate this one
	ImportPath string
	Action     string
	Output     string

	// The Action field is one of the following:
	// build-output - The toolchain printed output
	// build-fail - The build failed
}

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

func goBuild(ctx context.Context, path string) ([]BuildEvent, error) {
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
		return []BuildEvent{}, nil
	}

	var events []BuildEvent
	decoder := json.NewDecoder(bytes.NewReader(output))
	for {
		var event BuildEvent
		if err := decoder.Decode(&event); err == io.EOF {
			// TODO: Confirm the err is okay
			return events, nil
		} else if err != nil {
			return nil, err
		}
		events = append(events, event)
	}
}
