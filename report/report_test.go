package report_test

import (
	"github.com/ohhfishal/gopher/report"
	"github.com/ohhfishal/gopher/testdata"
	"github.com/stretchr/testify/assert"
	"io"
	"log/slog"
	"os"
	"path/filepath"
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

	tempDir := t.TempDir()
	for _, filename := range tests {
		t.Run(filename, func(t *testing.T) {
			assert := assert.New(t)
			file, err := testdata.FS.Open(filename)
			assert.Nil(err, "opening file")
			defer file.Close() //nolint: errcheck

			content, err := io.ReadAll(file)
			assert.Nil(err, "reading bytes")

			tempFile := filepath.Join(tempDir, filename)
			assert.Nil(os.WriteFile(tempFile, content, 0644))

			cmd := report.CMD{
				File: tempFile,
			}
			var stdout strings.Builder
			logger := slog.Default()
			err = cmd.Run(t.Context(), &stdout, logger)
			logger.Info("done", "err", err)
			assert.Nil(err)
		})
	}

}
