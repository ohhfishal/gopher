package watch

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

// TODO: Mention to the user to report this if they can
var ErrAssertion = errors.New("bug found: assertion failed")

func AssertFailed(reason string) error {
	return fmt.Errorf(`%w: %s`, ErrAssertion, reason)
}

type ReportCMD struct {
	FileContent []byte `short:"f" default:"-" type:"filecontent" help:"File to read from. Use '-' for stdin."`
}

func (config *ReportCMD) Run(ctx context.Context, stdout io.Writer, logger *slog.Logger) error {
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

type ErrorMessage interface {
	Print(io.Writer, int) error
	Add([]string) error
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
			if err := message.Print(stdout, 1); err != nil {
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

func (messages *ErrorMessages) Add(filename string, line string) error {
	if _, ok := messages.files[filename]; !ok {
		messages.files[filename] = map[string]ErrorMessage{}
	}
	errMap := messages.files[filename]

	parts := strings.Split(line, ":")
	if len(parts) < 3 {
		// TODO: Make a better message
		return AssertFailed("not enough parts: " + line)
	}

	errType := parts[2][1:]
	// Create a new message if the type does not exist
	if _, ok := errMap[errType]; !ok {
		var newMsg ErrorMessage
		switch errType {
		case "too many errors":
			messages.tooMany = true
			return nil
		case "undefined":
			newMsg = NewUndefinedErrorHandler()
		default:
			return fmt.Errorf("unknown error type: %s", errType)
		}
		errMap[errType] = newMsg
	}

	if err := errMap[errType].Add(parts); err != nil {
		return fmt.Errorf("adding: %s: %w", errType, err)
	}
	return nil
}

func aggregateErrors(importPath string, outputs []string) (*ErrorMessages, error) {
	steps := strings.Split(importPath, "/")
	child := steps[len(steps)-1] + "/"

	errorMsgs := NewErrorMessages()
	for _, output := range outputs {
		// Filter files that restate import path
		if strings.HasPrefix(output, "#") {
			continue
		}

		pretty, ok := strings.CutPrefix(output, child)
		if !ok {
			return nil, AssertFailed(`output does not start with same package`)
		}

		colonIndex := strings.IndexRune(pretty, ':')
		if colonIndex == -1 {
			// TODO: Fix
			return nil, AssertFailed(`colon missing`)
		}

		filename := pretty[:colonIndex]
		line := pretty[colonIndex+2:]
		if err := errorMsgs.Add(filename, line); err != nil {
			// TODO: Fix
			return nil, err
		}
	}
	return &errorMsgs, nil
}
