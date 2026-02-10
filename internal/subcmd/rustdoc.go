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

type RustDocCmd struct {
	cdr *subcommands.Commander
}

func (c *RustDocCmd) Name() string     { return "rustdoc" }
func (c *RustDocCmd) Synopsis() string { return "Search Rust documentation on docs.rs." }
func (c *RustDocCmd) Usage() string {
	return `rustdoc [flags]:
  Search Rust documentation on docs.rs.
`
}

func (c *RustDocCmd) SetFlags(f *flag.FlagSet) {
	cdr := subcommands.NewCommander(f, "")

	cdr.Register(cdr.CommandsCommand(), "help")
	cdr.Register(cdr.FlagsCommand(), "help")
	cdr.Register(cdr.HelpCommand(), "help")
	cdr.Register(&RustDocSearchCmd{}, "search")
	cdr.Register(&RustDocReadCmd{}, "read")

	c.cdr = cdr
}

func (c *RustDocCmd) Execute(
	ctx context.Context,
	f *flag.FlagSet,
	args ...any,
) subcommands.ExitStatus {
	return c.cdr.Execute(ctx, args...)
}

type RustDocSearchCmd struct{}

func (c *RustDocSearchCmd) Name() string     { return "search" }
func (c *RustDocSearchCmd) Synopsis() string { return "Search Rust crates on docs.rs." }
func (c *RustDocSearchCmd) Usage() string {
	return `rustdoc search <query>:
  Search Rust crates on docs.rs.
`
}

func (c *RustDocSearchCmd) SetFlags(f *flag.FlagSet) {
}

func (c *RustDocSearchCmd) Execute(
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
	result, err := app.SearchRustDoc(httpcli, query)
	if err != nil {
		fmt.Println("Error:", err)
		return subcommands.ExitFailure
	}
	fmt.Println(result)

	return subcommands.ExitSuccess
}

type RustDocReadCmd struct {
	offset int
	limit  int
}

func (c *RustDocReadCmd) Name() string     { return "read" }
func (c *RustDocReadCmd) Synopsis() string { return "Read Rust crate documentation from docs.rs." }
func (c *RustDocReadCmd) Usage() string {
	return `rustdoc read <crate_url>:
  Read Rust crate documentation from docs.rs.
`
}

func (c *RustDocReadCmd) SetFlags(f *flag.FlagSet) {
	f.IntVar(&c.offset, "offset", 0, "Line offset to start reading from")
	f.IntVar(&c.limit, "limit", app.DefaultLinesPerPage, "Number of lines to read")
}

func (p *RustDocReadCmd) Execute(
	_ context.Context,
	f *flag.FlagSet,
	_ ...any,
) subcommands.ExitStatus {
	if f.NArg() < 1 {
		fmt.Println("Error: Missing crate URL.")
		return subcommands.ExitUsageError
	}

	crateURL := f.Arg(0)
	httpcli := infra.NewHttpClient()
	result, totalLines, hasMore, err := app.ReadRustDocPaged(httpcli, crateURL, p.offset, p.limit)
	if err != nil {
		fmt.Println("Error:", err)
		return subcommands.ExitFailure
	}

	startLine := p.offset + 1
	endLine := p.offset + len(strings.Split(strings.TrimSpace(result), "\n"))
	if strings.TrimSpace(result) == "" {
		endLine = p.offset
	}

	fmt.Printf(
		"Documentation for '%s' (Lines %d-%d of %d):\n",
		crateURL,
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
