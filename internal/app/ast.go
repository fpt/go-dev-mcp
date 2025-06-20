package app

import (
	"context"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"

	"golang.org/x/tools/go/ast/astutil"

	"github.com/fpt/go-dev-mcp/internal/repository"
)

type FunctionExtractResult struct {
	Filename  string
	Functions []string
}

type FunctionCall struct {
	Name    string // Function name (e.g., "fmt.Println", "myFunc")
	Package string // Package name if external call (e.g., "fmt")
}

type CallGraphEntry struct {
	Function string         // Function name that makes calls
	Calls    []FunctionCall // Functions that it calls
}

type CallGraphResult struct {
	Filename  string           // Source file path
	CallGraph []CallGraphEntry // Call relationships
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

// ExtractCallGraph extracts function call relationships from a single Go file.
func ExtractCallGraph(filePath string) (*CallGraphResult, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	result := &CallGraphResult{
		Filename:  filePath,
		CallGraph: []CallGraphEntry{},
	}

	// Find all function declarations
	ast.Inspect(node, func(n ast.Node) bool {
		if funcDecl, ok := n.(*ast.FuncDecl); ok {
			if funcDecl.Name != nil && funcDecl.Name.IsExported() {
				entry := CallGraphEntry{
					Function: getFunctionSignature(funcDecl),
					Calls:    extractFunctionCalls(funcDecl),
				}
				result.CallGraph = append(result.CallGraph, entry)
			}
		}
		return true
	})

	return result, nil
}

// getFunctionSignature returns the function name with receiver if it's a method
func getFunctionSignature(funcDecl *ast.FuncDecl) string {
	if funcDecl.Recv != nil {
		receiverType := getReceiverType(funcDecl.Recv)
		return receiverType + "." + funcDecl.Name.Name
	}
	return funcDecl.Name.Name
}

// extractFunctionCalls finds all function calls within a function
func extractFunctionCalls(funcDecl *ast.FuncDecl) []FunctionCall {
	callMap := make(map[string]FunctionCall) // Use map to avoid duplicates

	if funcDecl.Body == nil {
		return []FunctionCall{}
	}

	ast.Inspect(funcDecl.Body, func(n ast.Node) bool {
		if callExpr, ok := n.(*ast.CallExpr); ok {
			call := analyzeFunctionCall(callExpr)
			if call.Name != "" {
				key := call.Package + "." + call.Name
				callMap[key] = call
			}
		}
		return true
	})

	// Convert map to slice with pre-allocation
	calls := make([]FunctionCall, 0, len(callMap))
	for _, call := range callMap {
		calls = append(calls, call)
	}

	return calls
}

// analyzeFunctionCall extracts function name and package from a call expression
func analyzeFunctionCall(callExpr *ast.CallExpr) FunctionCall {
	switch fun := callExpr.Fun.(type) {
	case *ast.Ident:
		// Local function call (e.g., "myFunc()")
		return FunctionCall{
			Name:    fun.Name,
			Package: "",
		}
	case *ast.SelectorExpr:
		// Package function call (e.g., "fmt.Println()")
		if ident, ok := fun.X.(*ast.Ident); ok {
			return FunctionCall{
				Name:    fun.Sel.Name,
				Package: ident.Name,
			}
		}
	case *ast.FuncLit:
		// Anonymous function - skip
		return FunctionCall{}
	}

	return FunctionCall{}
}
