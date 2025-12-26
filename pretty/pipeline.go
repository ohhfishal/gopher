package pretty

import (
	"io"
	"slices"
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
	if !pipeline.inline {
		if _, err := io.WriteString(pipeline.stdout, pipeline.delim); err != nil {
			return 0, err
		}
		pipeline.inline = true
	}

	if slices.Contains(content, '\n') {
		pipeline.inline = false
	}
	return pipeline.stdout.Write(content)
}
