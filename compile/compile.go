package compile

import (
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"os"
	"path/filepath"
)

type Target struct {
	Name        string
	Description string
}

func Compile(content []byte, dir string) error {
	if err := os.Mkdir(dir, 0750); err != nil && !os.IsExist(err) {
		return fmt.Errorf("making working directory: %w", err)
	}

	if err := os.WriteFile(filepath.Join(dir, "targets.go"), content, 0660); err != nil {
		return fmt.Errorf("copying over file: %w", err)
	}
	var err error
	var targets []Target

	targets, err = parseTargets(content)
	if err != nil {
		return fmt.Errorf("parsing targets: %w", err)
	} else if len(targets) == 0 {
		return fmt.Errorf("must include at least one target: %w", err)
	}

	var writer io.Writer
	if err = writeMain(writer, targets); err != nil {
		err = fmt.Errorf("writing main: %w", err)
		goto cleanup
	}

	// TODO: Actually build the binary
	return nil

cleanup:
	// Remove .gopher/*.go and delete the binary? Maybe just leave it?
	return fmt.Errorf("not implemented: cleanup: %w", err)
	return err
}

func writeMain(writer io.Writer, targets []Target) error {
	return errors.New("not implemented: main")
}

func parseTargets(content []byte) ([]Target, error) {
	tree, err := parser.ParseFile(
		token.NewFileSet(),
		"targets.go",
		content,
		parser.ParseComments,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to parse file: %w", err)
	}

	var targets []Target
	for _, decl := range tree.Decls {
		node, ok := decl.(*ast.FuncDecl)
		if !ok || !node.Name.IsExported() || !isValidFunc(node) {
			continue
		}

		targets = append(targets, Target{
			Name:        node.Name.Name,
			Description: "TODO IMPLEMENT PARSING COMMENT",
		})
	}
	return targets, nil
}

func isValidFunc(fn *ast.FuncDecl) bool {
	if fn.Type.Params == nil || fn.Type.Params.NumFields() != 2 {
		return false
	}

	if fn.Type.Results == nil || fn.Type.Results.NumFields() != 1 {
		return false
	}
	// TODO: Do more validation to make this more robust
	return true
}
