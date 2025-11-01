package report

import (
	"fmt"
	"io"
	"strings"
)

var stdlibImports = map[string]any{
	// TODO: Maybe make this check the go.mod too?
	"fmt":           nil,
	"io":            nil,
	"errors":        nil,
	"encoding/json": nil,
}

var errTypeMissingPackage = "missing package"

type ErrorMessage interface {
	Print(io.Writer) error
	Add([]string) error
}

type defaultErrorHandler struct {
	lines []string
}

func NewDefaultErrorHandler() ErrorMessage {
	return &defaultErrorHandler{}
}

func (h *defaultErrorHandler) Print(stdout io.Writer) error {
	for _, line := range h.lines {
		if _, err := fmt.Fprintln(stdout, " ", line); err != nil {
			return err
		}
	}
	return nil
}

func (h *defaultErrorHandler) Add(parts []string) error {
	h.lines = append(h.lines, strings.Join(parts, ":"))
	return nil
}

type undefinedErrorHandler struct {
	symbols   []string
	locations map[string][]Location
}

func NewUndefinedErrorHandler() ErrorMessage {
	return &undefinedErrorHandler{
		locations: map[string][]Location{},
	}
}
func (h *undefinedErrorHandler) Print(stdout io.Writer) error {
	if _, err := fmt.Fprintln(stdout, " ", "undefined:"); err != nil {
		return err
	}

	var potentialImports []string
	for _, symbol := range h.symbols {
		if _, ok := stdlibImports[symbol]; ok {
			potentialImports = append(potentialImports, symbol)
		}
		if _, err := fmt.Fprintln(stdout, "   ", symbol, LocationsString(h.locations[symbol])); err != nil {
			return err
		}
	}
	if len(potentialImports) > 0 {
		line := fmt.Sprintf(`Did you forget to import ("%s")?`, strings.Join(potentialImports, `", "`))
		if _, err := Suggestion.Fprintln(stdout, " ", line); err != nil {
			return err
		}
	}
	return nil
}
func (h *undefinedErrorHandler) Add(parts []string) error {
	if len(parts) != 4 {
		return fmt.Errorf("not enough parts (4): %s", strings.Join(parts, ":"))
	} else if parts[2] != " undefined" {
		return fmt.Errorf(`invalid lined must have "undefined": %s`, strings.Join(parts, ":"))
	} else if len(parts[3]) <= 1 {
		return fmt.Errorf(`unknown symbol: "%s"`, parts[3])
	}

	symbol := parts[3][1:]
	if _, ok := h.locations[symbol]; !ok {
		h.symbols = append(h.symbols, symbol)
		h.locations[symbol] = []Location{}
	}
	h.locations[symbol] = append(h.locations[symbol], Location{parts[0], parts[1]})

	return nil
}

type Location struct {
	Line string
	Col  string
}

func LocationsString(locations []Location) string {
	var truncated bool
	if len(locations) > 2 {
		locations = locations[:2]
		truncated = true
	}
	sliced := []string{}
	for _, loc := range locations {
		sliced = append(sliced, fmt.Sprintf("(%s:%s)", loc.Line, loc.Col))
	}
	body := strings.Join(sliced, ", ")
	if truncated {
		return fmt.Sprintf("%s, ...", body)
	}
	return body

}
