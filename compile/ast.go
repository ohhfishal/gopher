package compile

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
)

var aliases = map[string]string{
	"*Gopher": "*runtime.Gopher",
}

type funcSignature struct {
	Name       string
	Parameters []string
	Returns    []string
}

func (fn *funcSignature) String() string {
	retVal := strings.Join(fn.Returns, ", ")
	if len(fn.Returns) > 1 {
		retVal = fmt.Sprintf("(%s)", retVal)
	}
	return fmt.Sprintf("%s(%s) %s", fn.Name, strings.Join(fn.Parameters, ", "), retVal)
}

func fromFuncDecl(fn *ast.FuncDecl) funcSignature {
	returns := []string{}
	if rets := fn.Type.Results; rets != nil {
		for _, ret := range rets.List {
			returns = append(returns, getType(ret.Type))
		}
	}

	parameters := []string{}
	if params := fn.Type.Params; params != nil {
		for _, param := range params.List {
			t := getType(param.Type)
			if alias, ok := aliases[t]; ok {
				t = alias
			}
			parameters = append(parameters, t)
		}
	}

	signature := funcSignature{
		Name:       fn.Name.String(),
		Parameters: parameters,
		Returns:    returns,
	}
	return signature
}

func parseTargets(content []byte) ([]Target, []error, error) {
	tree, err := parser.ParseFile(
		token.NewFileSet(),
		TargetsFile,
		content,
		parser.ParseComments,
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse file: %w", err)
	}

	targets := []Target{}
	warnings := []error{}
	for _, decl := range tree.Decls {
		node, ok := decl.(*ast.FuncDecl)
		if !ok || !node.Name.IsExported() {
			continue
		} else if err := isValidFunc(node); err != nil {
			warnings = append(warnings, err)
			continue
		}

		comment := "No target description provided."
		if node.Doc != nil {
			if len(node.Doc.List) >= 1 {
				comment = NormalizeComment(node.Doc.List[0].Text)
				if len(node.Doc.List) > 1 {
					warnings = append(warnings,
						fmt.Errorf("unhandled edgecase: multi-comment doc comment for targets. Using only the first"),
					)
				}
			}
		}

		targets = append(targets, Target{
			Name:        node.Name.Name,
			Description: comment,
		})
	}
	return targets, warnings, nil
}

func isValidFunc(fn *ast.FuncDecl) error {
	expected := funcSignature{
		Name:       fn.Name.String(),
		Parameters: []string{"context.Context", "*runtime.Gopher"},
		Returns:    []string{"error"},
	}
	signature := fromFuncDecl(fn)

	errTail := fmt.Errorf("\n\t  have: %s\n\t  want: %s", signature.String(), expected.String())

	if fn.Type.TypeParams != nil && fn.Type.TypeParams.NumFields() != 0 {
		return fmt.Errorf("expected 0 type parameters got: %d", fn.Type.TypeParams.NumFields())
	} else if len(signature.Parameters) != len(expected.Parameters) {
		return fmt.Errorf("expected %d parameters got: %d\n%w",
			len(expected.Parameters),
			len(signature.Parameters),
			errTail,
		)
	} else if len(signature.Returns) != len(expected.Returns) {
		return fmt.Errorf("expected %d return value got: %d\n%w",
			len(expected.Returns),
			len(signature.Returns),
			errTail,
		)
	}

	for i, param := range signature.Parameters {
		expectedParam := expected.Parameters[i]
		if param != expectedParam {
			return fmt.Errorf("param %d: expected %s: got: %s\n%w", i, expectedParam, param, errTail)
		}
	}

	for i, ret := range signature.Returns {
		expectedRet := expected.Returns[i]
		if ret != expectedRet {
			return fmt.Errorf("ret %d: expected %s: got: %s\n%w", i, expectedRet, ret, errTail)
		}
	}
	return nil

}

func getType(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.ArrayType:
		return "[]" + getType(t.Elt)
	case *ast.StarExpr:
		return "*" + getType(t.X)
	case *ast.SelectorExpr:
		return getType(t.X) + "." + t.Sel.Name
	case *ast.MapType:
		return "map[" + getType(t.Key) + "]" + getType(t.Value)
	case *ast.ChanType:
		return "chan " + getType(t.Value)
	case *ast.InterfaceType:
		return "interface{}"
	case *ast.StructType:
		return "struct{}"
	case *ast.FuncType:
		return "func"
	default:
		return fmt.Sprintf("%T", expr)
	}
}
