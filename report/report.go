package report

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/fatih/color"
	"io"
	"log/slog"
	"strings"
)

var (
	Success    = color.New(color.FgGreen)
	Error      = color.New(color.FgRed)
	Suggestion = color.New(color.FgHiBlack)
)

const (
	BuildOutput = "build-output"
	BuildFail   = "build-fail"
)

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

// TODO: Mention to the user to report this if they can
var ErrAssertion = errors.New("bug found: assertion failed")

func AssertFailed(reason string) error {
	return fmt.Errorf(`%w: %s`, ErrAssertion, reason)
}

type CMD struct {
	FileContent []byte `short:"f" default:"-" type:"filecontent" help:"File to read from. Use '-' for stdin."`
}

func (config *CMD) Run(ctx context.Context, stdout io.Writer, logger *slog.Logger) error {
	events, err := ParseBuildJSON(bytes.NewReader(config.FileContent))
	if err != nil {
		return fmt.Errorf("parsing build output: %w", err)
	}

	if err := OutputResults(events, stdout); err != nil {
		return fmt.Errorf("parsing results: %w", err)
	}
	return nil
}

func OutputResults(events []BuildEvent, stdout io.Writer) error {
	if len(events) == 0 {
		_, err := Success.Fprintln(stdout, "OK")
		return err
	}

	// Assert preconditions
	if events[len(events)-1].Action != BuildFail {
		return AssertFailed(`events[-1].Action != "build-fail"`)
	}
	events = events[:len(events)-1]

	// Create a map for each package and errors
	var pathMap = map[string][]string{}
	for _, event := range events {
		if event.Action != BuildOutput {
			return AssertFailed(fmt.Sprintf(`events[:-1].Action != "build-output": "%s"`, event.Action))
		}
		event.Output, _ = strings.CutSuffix(event.Output, "\n")
		if outputs, ok := pathMap[event.ImportPath]; ok {
			pathMap[event.ImportPath] = append(outputs, event.Output)
		} else if !strings.HasPrefix(event.Output, "#") {
			// Exclude events that specify the import path, since we get it from JSON
			pathMap[event.ImportPath] = []string{event.Output}
		}
	}

	for importPath, outputs := range pathMap {
		errorMsgs, err := aggregateErrors(importPath, outputs)
		if err != nil {
			return err
		}

		if _, err := fmt.Fprintln(stdout, "package:", importPath); err != nil {
			return err
		}
		if err := errorMsgs.Print(stdout); err != nil {
			return err
		}
	}
	if _, err := Error.Fprintln(stdout, "FAILED"); err != nil {
		return err
	}
	return nil
}

func ParseBuildJSON(input io.Reader) ([]BuildEvent, error) {
	events := []BuildEvent{}
	decoder := json.NewDecoder(input)
	// TODO: Handle go vet output which on happy case looks like:
	// Looks like # is a comment
	/*
		# package
		# [package]
		{}
	*/
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

type ErrorMessages struct {
	files   map[string]map[string]ErrorMessage
	tooMany bool
}

func NewErrorMessages() ErrorMessages {
	return ErrorMessages{
		files: map[string]map[string]ErrorMessage{},
	}
}

func (mapping ErrorMessages) Print(stdout io.Writer) error {
	for file, errMessages := range mapping.files {
		if _, err := fmt.Fprintln(stdout, file); err != nil {
			return err
		}
		for _, message := range errMessages {
			if err := message.Print(stdout); err != nil {
				return err
			}
		}
	}
	if mapping.tooMany {
		if _, err := fmt.Fprintln(stdout, "\t", "..."); err != nil {
			return err
		}
	}
	return nil
}

func (messages *ErrorMessages) AddWithType(errType, filename string, line []string) error {
	if _, ok := messages.files[filename]; !ok {
		messages.files[filename] = map[string]ErrorMessage{}
	}
	errMap := messages.files[filename]

	// Create a new message if the type does not exist
	if _, ok := errMap[errType]; !ok {
		var newMsg ErrorMessage
		switch {
		case errType == errTypeMissingPackage:
			// TODO: Make a custom handler. Add adds go gets, then output a fancy line
			fallthrough
		case strings.HasPrefix(errType, "no required module provides package"):
			newMsg = NewDefaultErrorHandler()
		case errType == "too many errors":
			messages.tooMany = true
			return nil
		case errType == "undefined":
			newMsg = NewUndefinedErrorHandler()
		default:
			return fmt.Errorf(`unknown error type: "%s"`, errType)
		}
		errMap[errType] = newMsg
	}

	if err := errMap[errType].Add(line); err != nil {
		return fmt.Errorf("adding: %s: %w", errType, err)
	}
	return nil
}

func (messages *ErrorMessages) Add(filename string, line []string) error {
	errType := filename
	if len(line) >= 3 {
		errType = line[2][1:]
	}
	return messages.AddWithType(errType, filename, line)
}

func aggregateErrors(importPath string, outputs []string) (*ErrorMessages, error) {
	errorMsgs := NewErrorMessages()
	var filename string
	for _, output := range outputs {
		if strings.HasPrefix(output, "\tgo get") {
			if err := errorMsgs.AddWithType(
				errTypeMissingPackage,
				filename,
				[]string{output},
			); err != nil {
				// TODO: Fix
				return nil, err
			}
		} else if !strings.HasPrefix(output, "#") {
			parts := strings.Split(output, ":")
			if len(parts) < 4 {
				return nil, fmt.Errorf(`missing colons "%s"`, output)
			}
			filename = parts[0]
			line := parts[1:]

			if err := errorMsgs.Add(filename, line); err != nil {
				// TODO: Fix
				return nil, err
			}
		}
	}
	return &errorMsgs, nil
}
