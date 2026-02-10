package subcmd

import (
	"context"
	"flag"
	"fmt"

	"github.com/fpt/go-dev-mcp/internal/app"
	"github.com/fpt/go-dev-mcp/internal/infra"
	"github.com/google/subcommands"
)

type OutlineGoPackageCmd struct {
	directory        string
	skipDependencies bool
	skipDeclarations bool
	skipCallGraph    bool
}

func (*OutlineGoPackageCmd) Name() string { return "outline" }

func (*OutlineGoPackageCmd) Synopsis() string {
	return "Get a comprehensive outline of a Go package"
}

func (*OutlineGoPackageCmd) Usage() string {
	return `outline [flags] <directory>:
  Show dependencies, exported declarations, and call graph
  for all Go source files in the specified directory.
`
}

func (p *OutlineGoPackageCmd) SetFlags(f *flag.FlagSet) {
	f.StringVar(&p.directory, "dir", ".", "Directory containing Go source files")
	f.BoolVar(&p.skipDependencies, "skip-deps", false, "Skip the dependencies section")
	f.BoolVar(&p.skipDeclarations, "skip-decl", false, "Skip the declarations section")
	f.BoolVar(&p.skipCallGraph, "skip-cg", false, "Skip the call graph section")
}

func (p *OutlineGoPackageCmd) Execute(
	ctx context.Context,
	f *flag.FlagSet,
	_ ...any,
) subcommands.ExitStatus {
	directory := p.directory
	if f.NArg() > 0 {
		directory = f.Arg(0)
	}

	fw := infra.NewFileWalker()
	output, err := app.OutlineGoPackage(ctx, fw, directory, app.OutlineGoPackageOptions{
		SkipDependencies: p.skipDependencies,
		SkipDeclarations: p.skipDeclarations,
		SkipCallGraph:    p.skipCallGraph,
	})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return subcommands.ExitFailure
	}

	fmt.Print(output)
	return subcommands.ExitSuccess
}
