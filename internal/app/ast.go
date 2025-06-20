package app

import (
	"context"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"

	"golang.org/x/tools/go/ast/astutil"

	"fujlog.net/godev-mcp/internal/repository"
)

type FunctionExtractResult struct {
	Filename  string
	Functions []string
}

// ExtractFunctionNames extracts function names from Go source files in the specified directory.
func ExtractFunctionNames(
	ctx context.Context, fw repository.FileWalker, path string,
) ([]FunctionExtractResult, error) {
	var results []FunctionExtractResult
	err := fw.Walk(ctx, func(filePath string) error {
		functions := extractFunctionsFromFile(filePath)
		if len(functions) > 0 {
			results = append(results, FunctionExtractResult{
				Filename:  filePath,
				Functions: functions,
			})
		}

		return nil
	}, path, ".go", true)
	if err != nil {
		return nil, err
	}

	return results, nil
}

func extractFunctionsFromFile(filePath string) []string {
	// Skip test files
	if strings.HasSuffix(filePath, "_test.go") {
		return nil
	}

	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		// Skip files that can't be parsed instead of failing
		return nil
	}

	var functionNames []string
	astutil.Apply(node, nil, func(c *astutil.Cursor) bool {
		n := c.Node()
		switch x := n.(type) {
		case *ast.FuncDecl:
			if x.Name != nil && x.Name.IsExported() {
				// Extract both regular functions and methods
				if x.Recv != nil {
					// Method: include receiver type for clarity
					receiverType := getReceiverType(x.Recv)
					functionNames = append(functionNames, receiverType+"."+x.Name.Name)
				} else {
					// Regular function
					functionNames = append(functionNames, x.Name.Name)
				}
			}
		}
		return true
	})

	return functionNames
}

// getReceiverType extracts the receiver type name from a method receiver
func getReceiverType(recv *ast.FieldList) string {
	if recv == nil || len(recv.List) == 0 {
		return ""
	}

	field := recv.List[0]
	switch t := field.Type.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		if ident, ok := t.X.(*ast.Ident); ok {
			return "*" + ident.Name
		}
	}

	return "Unknown"
}
