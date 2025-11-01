package report_test

import (
	"github.com/ohhfishal/gopher/report"
	"github.com/ohhfishal/gopher/testdata"
	"github.com/stretchr/testify/assert"
	"io"
	"log/slog"
	"strings"
	"testing"
)

func TestExamples(t *testing.T) {
	tests := []string{
		"example1.txt",
		"example2.txt",
		"example3.txt",
		"example4.txt",
		"example5.txt",
	}

	for _, filename := range tests {
		t.Run(filename, func(t *testing.T) {
			assert := assert.New(t)
			file, err := testdata.FS.Open(filename)
			assert.Nil(err, "opening file")
			defer file.Close() //nolint: errcheck

			bytes, err := io.ReadAll(file)
			assert.Nil(err, "reading bytes")
			cmd := report.CMD{
				FileContent: bytes,
			}
			var stdout strings.Builder
			logger := slog.Default()
			err = cmd.Run(t.Context(), &stdout, logger)
			logger.Info("done", "err", err)
			assert.Nil(err)
		})
	}

}
