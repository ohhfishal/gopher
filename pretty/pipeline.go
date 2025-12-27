package pretty

import (
	"io"
	"strings"
)

var _ io.Writer = &Pipeline{}

type Pipeline struct {
	stdout io.Writer
	pos    int
	delim  string
	inline bool
}

func NewPipeline(stdout io.Writer, delim string) *Pipeline {
	return &Pipeline{
		stdout: stdout,
		delim:  delim,
	}
}

func (pipeline *Pipeline) Write(content []byte) (int, error) {
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
	left, right := contentStr[:index], contentStr[index:]
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
