package app

import (
	"context"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"

	"github.com/fpt/go-dev-mcp/internal/repository"
)

type Declaration struct {
	Name string
	Type string // "function", "type", "interface", "struct", "const", "var"
	Info string // Additional info like receiver type for methods, struct fields count, etc.
}

type DeclarationExtractResult struct {
	Filename     string
	Declarations []Declaration
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

// ExtractDeclarations extracts all exported declarations from Go source files in the specified directory.
func ExtractDeclarations(
	ctx context.Context, fw repository.FileWalker, path string,
) ([]DeclarationExtractResult, error) {
	var results []DeclarationExtractResult
	err := fw.Walk(ctx, func(filePath string) error {
		declarations := extractDeclarationsFromFile(filePath)
		if len(declarations) > 0 {
			results = append(results, DeclarationExtractResult{
				Filename:     filePath,
				Declarations: declarations,
			})
		}

		return nil
	}, path, ".go", true)
	if err != nil {
		return nil, err
	}

	return results, nil
}

// ExtractFunctionNames extracts function names from Go source files in the specified directory.
// This is kept for backward compatibility with existing MCP tools.
func ExtractFunctionNames(
	ctx context.Context, fw repository.FileWalker, path string,
) ([]DeclarationExtractResult, error) {
	results, err := ExtractDeclarations(ctx, fw, path)
	if err != nil {
		return nil, err
	}

	// Filter to only include function declarations for backward compatibility
	var functionResults []DeclarationExtractResult
	for _, result := range results {
		var functionDeclarations []Declaration
		for _, decl := range result.Declarations {
			if decl.Type == "function" {
				functionDeclarations = append(functionDeclarations, decl)
			}
		}
		if len(functionDeclarations) > 0 {
			functionResults = append(functionResults, DeclarationExtractResult{
				Filename:     result.Filename,
				Declarations: functionDeclarations,
			})
		}
	}

	return functionResults, nil
}

func extractDeclarationsFromFile(filePath string) []Declaration {
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

	var declarations []Declaration

	// Walk through all declarations in the file
	for _, decl := range node.Decls {
		switch d := decl.(type) {
		case *ast.GenDecl:
			// Handle type, const, var declarations
			for _, spec := range d.Specs {
				switch s := spec.(type) {
				case *ast.TypeSpec:
					if s.Name.IsExported() {
						decl := Declaration{
							Name: s.Name.Name,
							Type: getTypeSpecType(s),
							Info: getTypeSpecInfo(s),
						}
						declarations = append(declarations, decl)
					}
				case *ast.ValueSpec:
					// Handle const and var declarations
					declType := "var"
					if d.Tok == token.CONST {
						declType = "const"
					}
					for _, name := range s.Names {
						if name.IsExported() {
							decl := Declaration{
								Name: name.Name,
								Type: declType,
								Info: getValueSpecInfo(s),
							}
							declarations = append(declarations, decl)
						}
					}
				}
			}
		case *ast.FuncDecl:
			// Handle function declarations
			if d.Name != nil && d.Name.IsExported() {
				name := d.Name.Name
				info := ""
				if d.Recv != nil {
					// Method: include receiver type for clarity
					receiverType := getReceiverType(d.Recv)
					name = receiverType + "." + d.Name.Name
					info = "method on " + receiverType
				} else {
					info = "function"
				}
				decl := Declaration{
					Name: name,
					Type: "function",
					Info: info,
				}
				declarations = append(declarations, decl)
			}
		}
	}

	return declarations
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

// getTypeSpecType determines the specific type of a TypeSpec (struct, interface, type alias, etc.)
func getTypeSpecType(spec *ast.TypeSpec) string {
	switch spec.Type.(type) {
	case *ast.StructType:
		return "struct"
	case *ast.InterfaceType:
		return "interface"
	default:
		return "type"
	}
}

// getTypeSpecInfo provides additional information about a TypeSpec
func getTypeSpecInfo(spec *ast.TypeSpec) string {
	switch t := spec.Type.(type) {
	case *ast.StructType:
		fieldCount := len(t.Fields.List)
		if fieldCount == 1 {
			return "1 field"
		}
		return fmt.Sprintf("%d fields", fieldCount)
	case *ast.InterfaceType:
		methodCount := len(t.Methods.List)
		if methodCount == 1 {
			return "1 method"
		}
		return fmt.Sprintf("%d methods", methodCount)
	case *ast.Ident:
		return "alias to " + t.Name
	case *ast.ArrayType:
		return "array type"
	case *ast.MapType:
		return "map type"
	case *ast.ChanType:
		return "channel type"
	case *ast.FuncType:
		return "function type"
	default:
		return "custom type"
	}
}

// getValueSpecInfo provides additional information about a ValueSpec (const/var)
func getValueSpecInfo(spec *ast.ValueSpec) string {
	if spec.Type != nil {
		// Has explicit type
		return getTypeName(spec.Type)
	}
	if len(spec.Values) > 0 {
		// Infer from value
		return "inferred type"
	}
	return "no type info"
}

// getTypeName extracts a readable type name from an ast.Expr
func getTypeName(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.SelectorExpr:
		if pkg, ok := t.X.(*ast.Ident); ok {
			return pkg.Name + "." + t.Sel.Name
		}
		return "qualified type"
	case *ast.ArrayType:
		return "[]" + getTypeName(t.Elt)
	case *ast.MapType:
		return "map[" + getTypeName(t.Key) + "]" + getTypeName(t.Value)
	case *ast.StarExpr:
		return "*" + getTypeName(t.X)
	case *ast.ChanType:
		return "chan " + getTypeName(t.Value)
	default:
		return "complex type"
	}
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
