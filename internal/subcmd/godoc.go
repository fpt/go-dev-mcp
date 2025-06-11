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

type GoDocCmd struct {
	cdr *subcommands.Commander
}

func (c *GoDocCmd) Name() string     { return "godoc" }
func (c *GoDocCmd) Synopsis() string { return "Search Go documentation." }
func (c *GoDocCmd) Usage() string {
	return `godoc [flags]:
  Search Go documentation.
`
}

func (c *GoDocCmd) SetFlags(f *flag.FlagSet) {
	cdr := subcommands.NewCommander(f, "")

	cdr.Register(cdr.CommandsCommand(), "help")
	cdr.Register(cdr.FlagsCommand(), "help")
	cdr.Register(cdr.HelpCommand(), "help")
	cdr.Register(&GoDocSearchCmd{}, "search")
	cdr.Register(&GoDocReadCmd{}, "read")

	c.cdr = cdr
}

func (c *GoDocCmd) Execute(
	ctx context.Context,
	f *flag.FlagSet,
	args ...any,
) subcommands.ExitStatus {
	return c.cdr.Execute(ctx, args...)
}

// GoDocSearchCmd is a subcommand for searching Go documentation.
type GoDocSearchCmd struct{}

func (c *GoDocSearchCmd) Name() string     { return "search" }
func (c *GoDocSearchCmd) Synopsis() string { return "Search Go documentation." }
func (c *GoDocSearchCmd) Usage() string {
	return `godoc [flags]:
  Search Go documentation.
`
}

func (c *GoDocSearchCmd) SetFlags(f *flag.FlagSet) {
}

func (c *GoDocSearchCmd) Execute(
	_ context.Context,
	f *flag.FlagSet,
	_ ...any,
) subcommands.ExitStatus {
	if f.NArg() < 1 {
		fmt.Println("Error: Missing search query.")
		return subcommands.ExitUsageError
	}

	query := f.Arg(0)
	if query == "" {
		fmt.Println("Error: Empty search query.")
		return subcommands.ExitUsageError
	}

	httpcli := infra.NewHttpClient()
	result, err := app.SearchGoDoc(httpcli, query)
	if err != nil {
		fmt.Println("Error:", err)
		return subcommands.ExitFailure
	}
	fmt.Println(result)

	return subcommands.ExitSuccess
}

// GoDocReadCmd is a subcommand for reading Go documentation.
type GoDocReadCmd struct {
	offset int
	limit  int
}

func (c *GoDocReadCmd) Name() string     { return "read" }
func (c *GoDocReadCmd) Synopsis() string { return "Read Go documentation." }
func (c *GoDocReadCmd) Usage() string {
	return `godoc [flags]:
  Read Go documentation.
`
}

func (c *GoDocReadCmd) SetFlags(f *flag.FlagSet) {
	f.IntVar(&c.offset, "offset", 0, "Line offset to start reading from")
	f.IntVar(&c.limit, "limit", 50, "Number of lines to read")
}

func (p *GoDocReadCmd) Execute(
	_ context.Context,
	f *flag.FlagSet,
	_ ...any,
) subcommands.ExitStatus {
	if f.NArg() < 1 {
		fmt.Println("Error: Missing package URL.")
		return subcommands.ExitUsageError
	}

	packageURL := f.Arg(0)
	httpcli := infra.NewHttpClient()
	result, totalLines, hasMore, err := app.ReadGoDocPaged(httpcli, packageURL, p.offset, p.limit)
	if err != nil {
		fmt.Println("Error:", err)
		return subcommands.ExitFailure
	}

	// Calculate line range for display
	startLine := p.offset + 1
	endLine := p.offset + len(strings.Split(strings.TrimSpace(result), "\n"))
	if strings.TrimSpace(result) == "" {
		endLine = p.offset
	}

	fmt.Printf(
		"Documentation for '%s' (Lines %d-%d of %d):\n",
		packageURL,
		startLine,
		endLine,
		totalLines,
	)
	fmt.Println(result)

	if hasMore {
		nextOffset := p.offset + p.limit
		fmt.Printf("\n... (use --offset=%d to see more)\n", nextOffset)
	}

	return subcommands.ExitSuccess
}
