package pretty

import (
	"io"
	"strings"
)

var _ io.Writer = &IndentedWriter{}

type IndentedWriter struct {
	stdout io.Writer
	pos    int
	delim  string
	inline bool
}

func NewIndentedWriter(stdout io.Writer, delim string) *IndentedWriter {
	return &IndentedWriter{
		stdout: stdout,
		delim:  delim,
	}
}

func (pipeline *IndentedWriter) Write(content []byte) (int, error) {
	if len(content) == 0 {
		return 0, nil
	} else if !pipeline.inline {
		if _, err := io.WriteString(pipeline.stdout, pipeline.delim); err != nil {
			return 0, err
		}
		pipeline.inline = true
	}

	contentStr := string(content)
	index := strings.Index(contentStr, "\n")
	if index == -1 {
		// No newline
		pipeline.inline = true
		return pipeline.stdout.Write(content)
	} else if index == len(contentStr)-1 {
		// Single newline at the end
		pipeline.inline = false
		return pipeline.stdout.Write(content)
	}
	// Write the first line then recurse
	left, right := contentStr[:index+1], contentStr[index+1:]
	count, err := io.WriteString(pipeline.stdout, left)
	if err != nil {
		return count, err
	}
	pipeline.inline = false
	delta, err := pipeline.Write([]byte(right))
	if err != nil {
		return count, err
	}
	return count + delta, nil
}
