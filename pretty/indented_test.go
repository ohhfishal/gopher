package pretty_test

import (
	"io"
	"strings"
	"testing"

	"github.com/ohhfishal/gopher/pretty"
)

func TestIndentWriter(t *testing.T) {
	tests := []struct {
		Input  string
		Output string
	}{
		{Input: "", Output: ""},
		{Input: "\n", Output: "  \n"},
		{Input: "Hello World", Output: "  Hello World"},
		{Input: "\nHello World", Output: "  \n  Hello World"},
		{Input: "Hello World\n", Output: "  Hello World\n"},
		{Input: "Hello\nWorld", Output: "  Hello\n  World"},
		{Input: "\nHello World\n", Output: "  \n  Hello World\n"},
		{Input: "\nHello\nWorld", Output: "  \n  Hello\n  World"},
		{Input: "Hello\nWorld\n", Output: "  Hello\n  World\n"},
		{Input: "\nHello\nWorld\n", Output: "  \n  Hello\n  World\n"},
	}

	for _, test := range tests {
		for i := range len(test.Input) {
			t.Run(Raw(test.Input), func(t *testing.T) {
				left := test.Input[0:i]
				right := test.Input[i:]
				var builder strings.Builder
				writer := pretty.NewIndentedWriter(&builder, "  ")
				if _, err := io.WriteString(writer, left); err != nil {
					t.Fatalf("got error: %s", err.Error())
				} else if _, err := io.WriteString(writer, right); err != nil {
					t.Fatalf("got error: %s", err.Error())
				} else if output := builder.String(); output != test.Output {
					t.Fatalf(`got: "%s" expected: "%s"`, Raw(output), Raw(test.Output))
				}
			})
		}
	}
}

func Raw(str string) string {
	return strings.ReplaceAll(str, "\n", "\\n")
}
