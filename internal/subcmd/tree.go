package subcmd

import (
	"context"
	"flag"
	"fmt"
	"strings"

	"fujlog.net/godev-mcp/internal/app"
	"fujlog.net/godev-mcp/internal/infra"
	"github.com/google/subcommands"
)

type TreeCmd struct {
	workdir string
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
}

func (p *TreeCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...any) subcommands.ExitStatus {
	rootDir := p.workdir
	b := strings.Builder{}
	b.WriteString(fmt.Sprintf("%s\n", rootDir))
	walker := infra.NewDirWalker()
	app.PrintTree(&b, walker, rootDir, false)

	result := b.String()
	fmt.Println(result)

	return subcommands.ExitSuccess
}
