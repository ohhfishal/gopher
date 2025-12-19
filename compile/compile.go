package compile

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"github.com/ohhfishal/gopher/runner"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

var funcMap = template.FuncMap{
	"lower": strings.ToLower,
}

//go:embed template.go.tmpl
var rawMainTemplate string
var mainTemplate = template.Must(template.New("main.go").
	Funcs(funcMap).
	Parse(rawMainTemplate))

const BinaryName = "target"
const TargetsFile = "targets.go"

type Target struct {
	Name        string
	Description string
}

func Compile(content []byte, dir string, goBin string) error {
	if err := os.Mkdir(dir, 0750); err != nil && !os.IsExist(err) {
		return fmt.Errorf("making working directory: %w", err)
	}

	if err := os.WriteFile(filepath.Join(dir, TargetsFile), content, 0660); err != nil {
		return fmt.Errorf("copying over file: %w", err)
	}
	var err error
	var targets []Target

	targets, err = parseTargets(content)
	if err != nil {
		return fmt.Errorf("parsing targets: %w", err)
	} else if len(targets) == 0 {
		return fmt.Errorf("must include at least one target: %v", targets)
	}
	slog.Debug("parsed targets", "count", len(targets), "targets", targets)

	mainFile, err := os.OpenFile(filepath.Join(dir, "main.go"), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		err = fmt.Errorf("opening main.go: %w", err)
		goto cleanup
	}
	defer mainFile.Close()

	err = writeMain(mainFile, targets)
	if err != nil {
		err = fmt.Errorf("writing main.go: %w", err)
		goto cleanup
	}

	err = buildBinary(dir, goBin)
	if err != nil {
		err = fmt.Errorf("building binary: %w", err)
		goto cleanup
	}
	// TODO: Write metadata to the cache file
	return nil

cleanup:
	// Remove .gopher/*.go and delete the binary? Maybe just leave it?
	// TODO: Consider removing this. We are using a cache file which gives us a way to validate an error stopped us last time
	return fmt.Errorf("not implemented: cleanup: %w", err)
	return err
}

func buildBinary(dir string, goBin string) error {
	builder := runner.GoBuild{
		Output:   BinaryName,
		Flags:    []string{"-C", dir},
		Packages: []string{"main.go", TargetsFile},
	}
	var output strings.Builder
	err := builder.Run(context.TODO(), runner.RunArgs{
		GoConfig: runner.GoConfig{
			GoBin: goBin,
		},
		Stdout: &output,
	})
	slog.Debug("built", "path", filepath.Join(dir, BinaryName), "output", output.String())
	if errors.Is(runner.ErrOK, err) {
		return nil
	}
	return err
}

type TemplateData struct {
	Targets []Target
}

func writeMain(writer io.Writer, targets []Target) error {
	return mainTemplate.Execute(writer, TemplateData{
		Targets: targets,
	})
}

func parseTargets(content []byte) ([]Target, error) {
	tree, err := parser.ParseFile(
		token.NewFileSet(),
		TargetsFile,
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
