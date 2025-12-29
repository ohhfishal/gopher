package runtime

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type GitHook string

const (
	GitPreCommit GitHook = "pre-commit"
)

func InstallGitHook(stdout io.Writer, hook GitHook, content string) error {
	// TODO: Maybe only write if the file **does not** exist
	if err := os.WriteFile(
		filepath.Join(".git", "hooks", (string)(hook)),
		[]byte(content),
		0755,
	); err != nil {
		return fmt.Errorf("writing file: %w", err)
	}
	if stdout != nil {
		fmt.Fprintf(stdout, "Wrote Hook \"%s\":\n%s", (string)(hook), content)
	}
	return nil
}
