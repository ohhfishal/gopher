package report

import (
	"fmt"
	"io"
	"strings"
)

// var errTypeMissingPackage = "missing package"
// var errTypePackageNotInStd = "package not in stdlib"

type ErrorHandler interface {
	Print(io.Writer) error
	Add([]string) error
}

type defaultErrorHandler struct {
	lines []string
}

func NewDefaultErrorHandler() ErrorHandler {
	return &defaultErrorHandler{}
}

func (h *defaultErrorHandler) Print(stdout io.Writer) error {
	return PrintLines(stdout, h.lines)
}

func PrintLines(stdout io.Writer, lines []string) error {
	for _, line := range lines {
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

type TransformHandler struct {
	transform func([]string) ([]string, error)
	lines     []string
}

func NewTransformHandler(transform func([]string) ([]string, error)) ErrorHandler {
	return &TransformHandler{
		transform: transform,
		lines:     []string{},
	}
}

func (h *TransformHandler) Print(stdout io.Writer) error {
	return PrintLines(stdout, h.lines)
}

func (h *TransformHandler) Add(parts []string) error {
	lines, err := h.transform(parts)
	if err != nil {
		return err
	}
	h.lines = append(h.lines, lines...)
	return nil
}

type SuggestionFunc func(map[string][]Location) ([]string, error)

type AggregateHandler struct {
	index   int
	message string
	// lines []string
	lines      map[string][]Location
	suggestion SuggestionFunc
}

func NewAggregateHandler(message string, index int, suggestion SuggestionFunc) ErrorHandler {
	return &AggregateHandler{
		index:      index,
		message:    message,
		lines:      map[string][]Location{},
		suggestion: suggestion,
	}
}

func (h *AggregateHandler) Print(stdout io.Writer) error {
	lines := []string{h.message}
	for line, locations := range h.lines {
		lines = append(lines, fmt.Sprintf("\t%s %s", line, LocationsString(locations)))
	}
	if err := PrintLines(stdout, lines); err != nil {
		return err
	}
	if h.suggestion != nil {
		lines, err := h.suggestion(h.lines)
		if err != nil {
			return err
		}
		return PrintLines(stdout, lines)
	}
	return nil
}

func (h *AggregateHandler) Add(parts []string) error {
	if len(parts) <= h.index {
		return fmt.Errorf("not enough parts: %d expected: %d(%v)", h.index+1, len(parts), parts)
	}
	key := strings.TrimSpace(parts[h.index])
	if _, ok := h.lines[key]; !ok {
		h.lines[key] = []Location{}
	}
	h.lines[key] = append(h.lines[key], Location{parts[0], parts[1]})
	return nil
}

func UndefinedSuggestion(mapping map[string][]Location) ([]string, error) {
	var potentialImports []string
	for symbol := range mapping {
		if packagePath, ok := stdlibPackages[symbol]; ok {
			potentialImports = append(potentialImports, packagePath)
		}
	}
	if len(potentialImports) > 0 {
		line := fmt.Sprintf(`Did you forget to import ("%s")?`, strings.Join(potentialImports, `", "`))
		return []string{Suggestion(line)}, nil
	}
	return []string{}, nil
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

type badReturnValuesHandler struct {
	lines []string
}

func (h *badReturnValuesHandler) Print(stdout io.Writer) error {
	if len(h.lines) != 3 {
		return fmt.Errorf("did not get all lines needed: %s", h.lines)
	}

	haves := strings.Split(
		h.lines[1][strings.IndexRune(h.lines[1], '(')+1:strings.IndexRune(h.lines[1], ')')],
		", ",
	)
	needs := strings.Split(
		h.lines[2][strings.IndexRune(h.lines[2], '(')+1:strings.IndexRune(h.lines[2], ')')],
		", ",
	)

	for i, str := range haves {
		if i >= len(needs) {
			haves[i] = ColorRemove(str)
		} else if needs[i] != str {
			haves[i] = ColorRemove(str)
		}
	}

	for i, str := range needs {
		if i >= len(haves) {
			needs[i] = ColorAdd(str)
		} else if haves[i] != str {
			needs[i] = ColorAdd(str)
		}
	}

	return PrintLines(stdout, []string{
		" " + h.lines[0],
		fmt.Sprintf("\t have(%s)", strings.Join(haves, ", ")),
		fmt.Sprintf("\t need(%s)", strings.Join(needs, ", ")),
	})
}

func (h *badReturnValuesHandler) Add(parts []string) error {
	h.lines = append(h.lines, strings.Join(parts, ":"))
	return nil
}
