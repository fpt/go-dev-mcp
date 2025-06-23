package subcmd

import (
	"context"
	"flag"
	"fmt"

	"github.com/fpt/go-dev-mcp/internal/app"
	"github.com/fpt/go-dev-mcp/internal/infra"
	"github.com/google/subcommands"
)

type MarkdownCmd struct {
	path string
}

func (*MarkdownCmd) Name() string     { return "markdown" }
func (*MarkdownCmd) Synopsis() string { return "Scan markdown files and extract headings." }
func (*MarkdownCmd) Usage() string {
	return `markdown [flags]:
  Scan markdown files in a directory or analyze a single markdown file to extract headings with line numbers.
`
}

func (p *MarkdownCmd) SetFlags(f *flag.FlagSet) {
	f.StringVar(
		&p.path,
		"path",
		".",
		"Directory path to scan for markdown files or path to a single markdown file",
	)
}

func (p *MarkdownCmd) Execute(
	ctx context.Context,
	f *flag.FlagSet,
	args ...interface{},
) subcommands.ExitStatus {
	fw := infra.NewFileWalker()
	results, err := app.ScanMarkdownFiles(ctx, fw, p.path)
	if err != nil {
		fmt.Printf("Error scanning markdown files: %v\n", err)
		return subcommands.ExitFailure
	}

	output := app.FormatMarkdownScanResult(results)
	fmt.Print(output)
	return subcommands.ExitSuccess
}
