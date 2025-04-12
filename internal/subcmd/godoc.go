package subcmd

import (
	"context"
	"flag"
	"fmt"

	"fujlog.net/godev-mcp/internal/app"
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

func (c *GoDocCmd) Execute(ctx context.Context, f *flag.FlagSet, args ...any) subcommands.ExitStatus {
	return c.cdr.Execute(ctx, args...)
}

// GoDocSearchCmd is a subcommand for searching Go documentation.
type GoDocSearchCmd struct {
}

func (c *GoDocSearchCmd) Name() string     { return "search" }
func (c *GoDocSearchCmd) Synopsis() string { return "Search Go documentation." }
func (c *GoDocSearchCmd) Usage() string {
	return `godoc [flags]:
  Search Go documentation.
`
}

func (c *GoDocSearchCmd) SetFlags(f *flag.FlagSet) {
}

func (p *GoDocSearchCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...any) subcommands.ExitStatus {
	if f.NArg() < 1 {
		fmt.Println("Error: Missing search query.")
		return subcommands.ExitUsageError
	}

	query := f.Arg(0)
	result, err := app.SearchGoDoc(query)
	if err != nil {
		fmt.Println("Error:", err)
		return subcommands.ExitFailure
	}
	fmt.Println(result)

	return subcommands.ExitSuccess
}

// GoDocReadCmd is a subcommand for reading Go documentation.
type GoDocReadCmd struct {
}

func (c *GoDocReadCmd) Name() string     { return "read" }
func (c *GoDocReadCmd) Synopsis() string { return "Read Go documentation." }
func (c *GoDocReadCmd) Usage() string {
	return `godoc [flags]:
  Read Go documentation.
`
}

func (c *GoDocReadCmd) SetFlags(f *flag.FlagSet) {
}

func (p *GoDocReadCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...any) subcommands.ExitStatus {
	if f.NArg() < 1 {
		fmt.Println("Error: Missing package URL.")
		return subcommands.ExitUsageError
	}

	query := f.Arg(0)
	result, err := app.ReadGoDoc(query)
	if err != nil {
		fmt.Println("Error:", err)
		return subcommands.ExitFailure
	}
	fmt.Println(result)

	return subcommands.ExitSuccess
}
