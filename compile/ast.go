package compile

import (
	"fmt"
	"go/ast"
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
