package subcmd

import (
	"context"
	"flag"
	"fmt"
	"strings"

	"github.com/fpt/go-dev-mcp/internal/app"
	"github.com/fpt/go-dev-mcp/internal/infra"
	"github.com/google/subcommands"
)

type TreeCmd struct {
	workdir   string
	ignoreDot bool
	maxDepth  int
}

func (*TreeCmd) Name() string     { return "tree" }
func (*TreeCmd) Synopsis() string { return "List files in the workspace directory." }
func (*TreeCmd) Usage() string {
	return `serve [flags]:
  Serve files over the server.
`
}

func (p *TreeCmd) SetFlags(f *flag.FlagSet) {
	f.StringVar(&p.workdir, "workdir", ".", "Working directory")
	f.BoolVar(
		&p.ignoreDot,
		"ignore-dot",
		false,
		"Ignore dot files and directories (except .git which is always ignored)",
	)
	f.IntVar(&p.maxDepth, "max-depth", 4, "Maximum depth for directory traversal")
}

func (p *TreeCmd) Execute(ctx context.Context, f *flag.FlagSet, _ ...any) subcommands.ExitStatus {
	rootDir := p.workdir
	b := strings.Builder{}
	b.WriteString(fmt.Sprintf("%s\n", rootDir))
	walker := infra.NewDirWalker()
	err := app.PrintTree(ctx, &b, walker, rootDir, p.ignoreDot, p.maxDepth)
	if err != nil {
		fmt.Printf("Error printing tree: %v\n", err)
		return subcommands.ExitFailure
	}

	result := b.String()
	fmt.Println(result)

	return subcommands.ExitSuccess
}
