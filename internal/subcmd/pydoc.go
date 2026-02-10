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

type PyDocCmd struct {
	cdr *subcommands.Commander
}

func (c *PyDocCmd) Name() string     { return "pydoc" }
func (c *PyDocCmd) Synopsis() string { return "Search Python documentation on docs.python.org." }
func (c *PyDocCmd) Usage() string {
	return `pydoc [flags]:
  Search Python documentation on docs.python.org.
`
}

func (c *PyDocCmd) SetFlags(f *flag.FlagSet) {
	cdr := subcommands.NewCommander(f, "")

	cdr.Register(cdr.CommandsCommand(), "help")
	cdr.Register(cdr.FlagsCommand(), "help")
	cdr.Register(cdr.HelpCommand(), "help")
	cdr.Register(&PyDocSearchCmd{}, "search")
	cdr.Register(&PyDocReadCmd{}, "read")

	c.cdr = cdr
}

func (c *PyDocCmd) Execute(
	ctx context.Context,
	f *flag.FlagSet,
	args ...any,
) subcommands.ExitStatus {
	return c.cdr.Execute(ctx, args...)
}

type PyDocSearchCmd struct{}

func (c *PyDocSearchCmd) Name() string     { return "search" }
func (c *PyDocSearchCmd) Synopsis() string { return "Search Python standard library modules." }
func (c *PyDocSearchCmd) Usage() string {
	return `pydoc search <query>:
  Search Python standard library modules on docs.python.org.
`
}

func (c *PyDocSearchCmd) SetFlags(f *flag.FlagSet) {
}

func (c *PyDocSearchCmd) Execute(
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
	result, err := app.SearchPyDoc(httpcli, query)
	if err != nil {
		fmt.Println("Error:", err)
		return subcommands.ExitFailure
	}
	fmt.Println(result)

	return subcommands.ExitSuccess
}

type PyDocReadCmd struct {
	offset int
	limit  int
}

func (c *PyDocReadCmd) Name() string { return "read" }
func (c *PyDocReadCmd) Synopsis() string {
	return "Read Python module documentation from docs.python.org."
}
func (c *PyDocReadCmd) Usage() string {
	return `pydoc read <module_name>:
  Read Python module documentation from docs.python.org.
`
}

func (c *PyDocReadCmd) SetFlags(f *flag.FlagSet) {
	f.IntVar(&c.offset, "offset", 0, "Line offset to start reading from")
	f.IntVar(&c.limit, "limit", app.DefaultLinesPerPage, "Number of lines to read")
}

func (p *PyDocReadCmd) Execute(
	_ context.Context,
	f *flag.FlagSet,
	_ ...any,
) subcommands.ExitStatus {
	if f.NArg() < 1 {
		fmt.Println("Error: Missing module name.")
		return subcommands.ExitUsageError
	}

	moduleName := f.Arg(0)
	httpcli := infra.NewHttpClient()
	result, totalLines, hasMore, err := app.ReadPyDocPaged(httpcli, moduleName, p.offset, p.limit)
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
		moduleName,
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
