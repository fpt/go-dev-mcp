package app

import (
	"context"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"github.com/fpt/go-dev-mcp/internal/repository"
)

type Declaration struct {
	Name string
	Type string // "function", "type", "interface", "struct", "const", "var"
	Info string // Additional info like receiver type for methods, struct fields count, etc.
	Line int    // Line number in the file where the declaration starts
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

// PackageImport represents an import statement in a Go file
type PackageImport struct {
	Path     string // Import path (e.g., "fmt", "github.com/user/repo/pkg")
	Alias    string // Import alias if any (e.g., "f" in `import f "fmt"`)
	Line     int    // Line number where import appears
	IsStdlib bool   // Whether this is a standard library import
	IsLocal  bool   // Whether this is a local project import
}

// PackageDependency represents a Go file's package and its imports
type PackageDependency struct {
	FilePath    string          // Absolute path to the Go file
	PackageName string          // Package name declared in the file
	Imports     []PackageImport // All imports in the file
}

// DependencyGraphResult represents the package dependency analysis for a directory
type DependencyGraphResult struct {
	ProjectPath  string              // Root project path
	ModuleName   string              // Module name from go.mod
	Dependencies []PackageDependency // Package dependencies for each file
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
							Line: fset.Position(s.Pos()).Line,
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
								Line: fset.Position(name.Pos()).Line,
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
					Line: fset.Position(d.Pos()).Line,
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

// ExtractPackageDependencies analyzes Go files in a directory and extracts package-level import dependencies
func ExtractPackageDependencies(
	ctx context.Context,
	fw repository.FileWalker,
	projectPath string,
) (*DependencyGraphResult, error) {
	result := &DependencyGraphResult{
		ProjectPath:  projectPath,
		Dependencies: []PackageDependency{},
	}

	// Try to find module name from go.mod
	moduleName, err := findModuleName(projectPath)
	if err != nil {
		// If no go.mod found, use directory name as fallback
		moduleName = filepath.Base(projectPath)
	}
	result.ModuleName = moduleName

	// Walk through all Go files and extract dependencies
	err = fw.Walk(ctx, func(filePath string) error {
		dep, err := extractPackageDependencyFromFile(filePath, moduleName)
		if err != nil {
			// Skip files that can't be parsed instead of failing
			return nil
		}

		if dep != nil {
			result.Dependencies = append(result.Dependencies, *dep)
		}

		return nil
	}, projectPath, ".go", true)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// findModuleName reads the module name from go.mod file
func findModuleName(projectPath string) (string, error) {
	goModPath := filepath.Join(projectPath, "go.mod")
	content, err := os.ReadFile(goModPath)
	if err != nil {
		return "", err
	}

	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "module")), nil
		}
	}

	return "", fmt.Errorf("module declaration not found in go.mod")
}

// extractPackageDependencyFromFile extracts package info and imports from a single Go file
func extractPackageDependencyFromFile(filePath, moduleName string) (*PackageDependency, error) {
	// Skip test files
	if strings.HasSuffix(filePath, "_test.go") {
		return nil, nil
	}

	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	absFilePath, err := filepath.Abs(filePath)
	if err != nil {
		absFilePath = filePath
	}

	dep := &PackageDependency{
		FilePath:    absFilePath,
		PackageName: node.Name.Name,
		Imports:     []PackageImport{},
	}

	// Extract import declarations
	for _, imp := range node.Imports {
		importPath := strings.Trim(imp.Path.Value, `"`)

		var alias string
		if imp.Name != nil {
			alias = imp.Name.Name
		}

		line := fset.Position(imp.Pos()).Line

		packageImport := PackageImport{
			Path:     importPath,
			Alias:    alias,
			Line:     line,
			IsStdlib: isStandardLibrary(importPath),
			IsLocal:  isLocalImport(importPath, moduleName),
		}

		dep.Imports = append(dep.Imports, packageImport)
	}

	return dep, nil
}

// isStandardLibrary checks if an import path is from Go's standard library
func isStandardLibrary(importPath string) bool {
	// Standard library packages don't contain dots or are well-known exceptions
	if !strings.Contains(importPath, ".") {
		return true
	}

	// Known standard library packages with dots
	stdlibWithDots := []string{
		"golang.org/x/",
		"vendor/golang.org/x/",
	}

	for _, prefix := range stdlibWithDots {
		if strings.HasPrefix(importPath, prefix) {
			return true
		}
	}

	return false
}

// isLocalImport checks if an import path is from the local project
func isLocalImport(importPath, moduleName string) bool {
	return strings.HasPrefix(importPath, moduleName)
}

// OutlineGoPackageOptions controls which sections are included in the outline.
type OutlineGoPackageOptions struct {
	SkipDependencies  bool
	SkipDeclarations  bool
	SkipCallGraph     bool
}

// OutlineGoPackage produces a comprehensive outline of a Go package:
// dependencies, exported declarations, and call graph.
// Individual sections can be disabled via opts to reduce output size.
func OutlineGoPackage(
	ctx context.Context, fw repository.FileWalker, directory string, opts OutlineGoPackageOptions,
) (string, error) {
	var sb strings.Builder

	// Always extract dependencies for the module name header
	depResult, err := ExtractPackageDependencies(ctx, fw, directory)
	if err != nil {
		return "", fmt.Errorf("extracting dependencies: %w", err)
	}

	sb.WriteString(fmt.Sprintf("Package outline for: %s\n", directory))
	sb.WriteString(fmt.Sprintf("Module: %s\n\n", depResult.ModuleName))

	if !opts.SkipDependencies {
		sb.WriteString("== Dependencies ==\n")
		formatDepsRelative(&sb, depResult)
	}

	if !opts.SkipDeclarations {
		declResults, err := ExtractDeclarations(ctx, fw, directory)
		if err != nil {
			return "", fmt.Errorf("extracting declarations: %w", err)
		}

		sb.WriteString("== Declarations ==\n")
		if len(declResults) == 0 {
			sb.WriteString("No exported declarations found.\n")
		} else {
			for _, result := range declResults {
				sb.WriteString(fmt.Sprintf("File: %s\n", result.Filename))
				for _, decl := range result.Declarations {
					if decl.Info != "" {
						sb.WriteString(fmt.Sprintf(
							"- %s: %s (%s) [line %d]\n",
							decl.Type, decl.Name, decl.Info, decl.Line,
						))
					} else {
						sb.WriteString(fmt.Sprintf("- %s: %s [line %d]\n", decl.Type, decl.Name, decl.Line))
					}
				}
			}
		}
		sb.WriteString("\n")
	}

	if !opts.SkipCallGraph {
		sb.WriteString("== Call Graph ==\n")
		callGraphCount := 0

		err = fw.Walk(ctx, func(filePath string) error {
			if strings.HasSuffix(filePath, "_test.go") {
				return nil
			}

			result, cgErr := ExtractCallGraph(filePath)
			if cgErr != nil {
				return nil // skip unparseable files
			}

			if len(result.CallGraph) == 0 {
				return nil
			}

			callGraphCount++
			sb.WriteString(fmt.Sprintf("File: %s\n", result.Filename))

			for _, entry := range result.CallGraph {
				calls := filterCallGraphNoise(entry.Calls)
				if len(calls) == 0 {
					continue
				}

				sb.WriteString(fmt.Sprintf("  %s\n", entry.Function))

				for _, call := range calls {
					if call.Package != "" {
						sb.WriteString(fmt.Sprintf("    -> %s.%s\n", call.Package, call.Name))
					} else {
						sb.WriteString(fmt.Sprintf("    -> %s\n", call.Name))
					}
				}
			}

			return nil
		}, directory, ".go", true)
		if err != nil {
			return "", fmt.Errorf("walking for call graph: %w", err)
		}

		if callGraphCount == 0 {
			sb.WriteString("No exported function calls found.\n")
		}
	}

	return sb.String(), nil
}

// formatDepsRelative writes dependency info using relative file paths.
func formatDepsRelative(sb *strings.Builder, result *DependencyGraphResult) {
	if len(result.Dependencies) == 0 {
		sb.WriteString("No Go files with imports found.\n")
		return
	}

	cwd, _ := os.Getwd()

	for _, dep := range result.Dependencies {
		if len(dep.Imports) == 0 {
			continue
		}

		relPath := dep.FilePath
		if cwd != "" {
			if r, err := filepath.Rel(cwd, dep.FilePath); err == nil {
				relPath = r
			}
		}

		sb.WriteString(fmt.Sprintf("%s (package %s)\n", relPath, dep.PackageName))

		var localImports, stdlibImports, externalImports []PackageImport
		for _, imp := range dep.Imports {
			if imp.IsLocal {
				localImports = append(localImports, imp)
			} else if imp.IsStdlib {
				stdlibImports = append(stdlibImports, imp)
			} else {
				externalImports = append(externalImports, imp)
			}
		}

		if len(localImports) > 0 {
			sb.WriteString("  Local: ")
			for i, imp := range localImports {
				if i > 0 {
					sb.WriteString(", ")
				}
				sb.WriteString(imp.Path)
			}
			sb.WriteString("\n")
		}

		if len(externalImports) > 0 {
			sb.WriteString("  External: ")
			for i, imp := range externalImports {
				if i > 0 {
					sb.WriteString(", ")
				}
				sb.WriteString(imp.Path)
			}
			sb.WriteString("\n")
		}

		if len(stdlibImports) > 0 {
			sb.WriteString("  Stdlib: ")
			for i, imp := range stdlibImports {
				if i > 0 {
					sb.WriteString(", ")
				}
				sb.WriteString(imp.Path)
			}
			sb.WriteString("\n")
		}
	}

	sb.WriteString("\n")
}

// callGraphNoisePackages contains stdlib packages whose calls are too common to be useful.
var callGraphNoisePackages = map[string]bool{
	"fmt": true, "strings": true, "strconv": true, "bytes": true,
	"sort": true, "slices": true, "maps": true,
	"log": true, "slog": true,
}

// callGraphNoiseBuiltins contains built-in functions that add no insight.
var callGraphNoiseBuiltins = map[string]bool{
	"len": true, "cap": true, "make": true, "new": true,
	"append": true, "copy": true, "delete": true, "close": true,
	"panic": true, "recover": true, "print": true, "println": true,
	"error": true,
}

// filterCallGraphNoise removes trivial calls (builtins, common stdlib,
// unexported locals, method calls on local variables) to keep the call graph
// focused on meaningful relationships.
func filterCallGraphNoise(calls []FunctionCall) []FunctionCall {
	var filtered []FunctionCall
	for _, c := range calls {
		// Skip builtins
		if c.Package == "" && callGraphNoiseBuiltins[c.Name] {
			continue
		}
		// Skip noisy stdlib packages
		if callGraphNoisePackages[c.Package] {
			continue
		}
		// Skip method calls on local variables (e.g., m.Match, sb.WriteString)
		// — the AST parser reports these as Package="m", Name="Match";
		// real package names are always lowercase but local vars are too,
		// so we distinguish by checking if the "package" looks like a variable
		// (no dots, not in the import list). A good heuristic: real Go package
		// identifiers are multi-char and conventionally not single letters.
		if c.Package != "" && len(c.Package) <= 2 {
			continue
		}
		// Skip unexported local function calls
		if c.Package == "" && len(c.Name) > 0 && c.Name[0] >= 'a' && c.Name[0] <= 'z' {
			continue
		}
		filtered = append(filtered, c)
	}
	return filtered
}

// FormatDependencyGraph formats the dependency graph results as a readable string
func FormatDependencyGraph(result *DependencyGraphResult) string {
	if len(result.Dependencies) == 0 {
		return "No Go files with imports found."
	}

	var output strings.Builder
	output.WriteString(fmt.Sprintf("Package Dependencies for module: %s\n", result.ModuleName))
	output.WriteString(fmt.Sprintf("Project path: %s\n\n", result.ProjectPath))

	for _, dep := range result.Dependencies {
		if len(dep.Imports) == 0 {
			continue // Skip files with no imports
		}

		output.WriteString(fmt.Sprintf("%s (package %s)\n", dep.FilePath, dep.PackageName))

		// Group imports by type
		var localImports, stdlibImports, externalImports []PackageImport
		for _, imp := range dep.Imports {
			if imp.IsLocal {
				localImports = append(localImports, imp)
			} else if imp.IsStdlib {
				stdlibImports = append(stdlibImports, imp)
			} else {
				externalImports = append(externalImports, imp)
			}
		}

		// Display local imports first
		if len(localImports) > 0 {
			output.WriteString("  Local imports:\n")
			for _, imp := range localImports {
				output.WriteString(fmt.Sprintf("    %s (line %d)\n", imp.Path, imp.Line))
				if imp.Alias != "" && imp.Alias != "." && imp.Alias != "_" {
					output.WriteString(fmt.Sprintf("      alias: %s\n", imp.Alias))
				}
			}
		}

		// Display external imports
		if len(externalImports) > 0 {
			output.WriteString("  External imports:\n")
			for _, imp := range externalImports {
				output.WriteString(fmt.Sprintf("    %s (line %d)\n", imp.Path, imp.Line))
				if imp.Alias != "" && imp.Alias != "." && imp.Alias != "_" {
					output.WriteString(fmt.Sprintf("      alias: %s\n", imp.Alias))
				}
			}
		}

		// Display stdlib imports
		if len(stdlibImports) > 0 {
			output.WriteString("  Standard library imports:\n")
			for _, imp := range stdlibImports {
				output.WriteString(fmt.Sprintf("    %s (line %d)\n", imp.Path, imp.Line))
				if imp.Alias != "" && imp.Alias != "." && imp.Alias != "_" {
					output.WriteString(fmt.Sprintf("      alias: %s\n", imp.Alias))
				}
			}
		}

		output.WriteString("\n")
	}

	return output.String()
}
