package report_test

import (
	"github.com/ohhfishal/gopher/report"
	"github.com/ohhfishal/gopher/testdata"
	"github.com/stretchr/testify/assert"
	"io"
	"io/fs"
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

func TestBuildOutputs(t *testing.T) {
	entries, err := fs.ReadDir(testdata.FS, "buildOutputs")
	if err != nil {
		t.Fatalf("failed to read buildOutputs directory: %v", err)
	}

	tempDir := t.TempDir()
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		t.Run(entry.Name(), func(t *testing.T) {
			assert := assert.New(t)

			file, err := testdata.FS.Open(filepath.Join("buildOutputs", entry.Name()))
			assert.Nil(err, "opening file")
			defer file.Close() //nolint: errcheck

			content, err := io.ReadAll(file)
			assert.Nil(err, "reading bytes")

			tempFile := filepath.Join(tempDir, entry.Name())
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
