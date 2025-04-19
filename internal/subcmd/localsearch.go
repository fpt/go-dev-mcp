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

type LocalSearchCmd struct {
	extension string
}

func (*LocalSearchCmd) Name() string     { return "localsearch" }
func (*LocalSearchCmd) Synopsis() string { return "Search for files locally." }
func (*LocalSearchCmd) Usage() string {
	return `localsearch [flags]:
  Search for files locally.
`
}

func (p *LocalSearchCmd) SetFlags(f *flag.FlagSet) {
	f.StringVar(&p.extension, "extension", "", "File extension to search for")
}

func (p *LocalSearchCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...any) subcommands.ExitStatus {
	fw := infra.NewFileWalker()
	path := f.Arg(0)
	query := f.Arg(1)
	if path == "" {
		fmt.Println("Path is required")
		return subcommands.ExitFailure
	}
	if query == "" {
		fmt.Println("Query is required")
		return subcommands.ExitFailure
	}

	extension := p.extension
	if extension != "" && !strings.HasPrefix(extension, ".") {
		extension = "." + extension
	}

	results, err := app.SearchLocalFiles(context.Background(), fw, path, extension, query)
	if err != nil {
		fmt.Printf("Error searching local files: %v\n", err)
		return subcommands.ExitFailure
	}

	if len(results) == 0 {
		fmt.Println("No results found")
		return subcommands.ExitSuccess
	}
	for _, result := range results {
		for _, match := range result.Matches {
			fmt.Printf("Found file: %s\nMatch: %s (Line: %d)\n",
				result.Filename, match.Text, match.LineNo,
			)
		}
		fmt.Println()
	}

	return subcommands.ExitSuccess
}
