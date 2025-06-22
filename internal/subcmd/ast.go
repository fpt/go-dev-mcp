package subcmd

import (
	"context"
	"flag"
	"fmt"

	"github.com/fpt/go-dev-mcp/internal/app"
	"github.com/fpt/go-dev-mcp/internal/infra"
	"github.com/google/subcommands"
)

type ExtractDeclarationsCmd struct {
	directory string
}

func (*ExtractDeclarationsCmd) Name() string { return "extract-declarations" }
func (*ExtractDeclarationsCmd) Synopsis() string {
	return "Extract exported declarations from Go files"
}

func (*ExtractDeclarationsCmd) Usage() string {
	return `extract-declarations [flags] <directory>:
  Extract exported declarations (functions, types, interfaces, structs, constants, variables) 
  from Go source files in the specified directory.
`
}

func (p *ExtractDeclarationsCmd) SetFlags(f *flag.FlagSet) {
	f.StringVar(&p.directory, "dir", ".", "Directory to search for Go files")
}

func (p *ExtractDeclarationsCmd) Execute(
	ctx context.Context,
	f *flag.FlagSet,
	_ ...any,
) subcommands.ExitStatus {
	directory := p.directory
	if f.NArg() > 0 {
		directory = f.Arg(0)
	}

	fw := infra.NewFileWalker()
	results, err := app.ExtractDeclarations(ctx, fw, directory)
	if err != nil {
		fmt.Printf("Error extracting declarations: %v\n", err)
		return subcommands.ExitFailure
	}

	if len(results) == 0 {
		fmt.Println("No Go declarations found in the specified directory.")
		return subcommands.ExitSuccess
	}

	for _, result := range results {
		fmt.Printf("File: %s\n", result.Filename)
		for _, decl := range result.Declarations {
			if decl.Info != "" {
				fmt.Printf("  %s: %s (%s)\n", decl.Type, decl.Name, decl.Info)
			} else {
				fmt.Printf("  %s: %s\n", decl.Type, decl.Name)
			}
		}
		fmt.Println()
	}

	return subcommands.ExitSuccess
}

type ExtractCallGraphCmd struct {
	filePath string
}

func (*ExtractCallGraphCmd) Name() string { return "extract-call-graph" }

func (*ExtractCallGraphCmd) Synopsis() string { return "Extract function call graph from a Go file" }
func (*ExtractCallGraphCmd) Usage() string {
	return `extract-call-graph [flags] <file>:
  Extract function call relationships from a single Go source file.
`
}

func (p *ExtractCallGraphCmd) SetFlags(f *flag.FlagSet) {
	f.StringVar(&p.filePath, "file", "", "Go source file to analyze")
}

func (p *ExtractCallGraphCmd) Execute(
	ctx context.Context,
	f *flag.FlagSet,
	_ ...any,
) subcommands.ExitStatus {
	filePath := p.filePath
	if f.NArg() > 0 {
		filePath = f.Arg(0)
	}

	if filePath == "" {
		fmt.Println("Error: file path is required")
		return subcommands.ExitFailure
	}

	result, err := app.ExtractCallGraph(filePath)
	if err != nil {
		fmt.Printf("Error extracting call graph: %v\n", err)
		return subcommands.ExitFailure
	}

	fmt.Printf("Call Graph for: %s\n\n", result.Filename)

	if len(result.CallGraph) == 0 {
		fmt.Println("No exported functions found in the file.")
		return subcommands.ExitSuccess
	}

	for _, entry := range result.CallGraph {
		fmt.Printf("Function: %s\n", entry.Function)
		if len(entry.Calls) == 0 {
			fmt.Println("  - No function calls")
		} else {
			for _, call := range entry.Calls {
				if call.Package != "" {
					fmt.Printf("  - %s.%s\n", call.Package, call.Name)
				} else {
					fmt.Printf("  - %s (local)\n", call.Name)
				}
			}
		}
		fmt.Println()
	}

	return subcommands.ExitSuccess
}
