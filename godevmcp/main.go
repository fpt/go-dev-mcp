package main

import (
	"context"
	"flag"
	"os"

	"github.com/fpt/go-dev-mcp/internal/subcmd"
	"github.com/google/subcommands"
)

func main() {
	subcommands.Register(subcommands.HelpCommand(), "")
	subcommands.Register(subcommands.FlagsCommand(), "")
	subcommands.Register(subcommands.CommandsCommand(), "")
	subcommands.Register(&subcmd.ServeCmd{}, "")
	subcommands.Register(&subcmd.TreeCmd{}, "")
	subcommands.Register(&subcmd.GoDocCmd{}, "")
	subcommands.Register(&subcmd.RustDocCmd{}, "")
	subcommands.Register(&subcmd.GithubCmd{}, "")
	subcommands.Register(&subcmd.LocalSearchCmd{}, "")
	subcommands.Register(&subcmd.OutlineGoPackageCmd{}, "")
	subcommands.Register(&subcmd.MarkdownCmd{}, "")
	subcommands.Register(&subcmd.ValidateCmd{}, "")
	subcommands.Register(&subcmd.PyDocCmd{}, "")

	flag.Parse()
	ctx := context.Background()
	os.Exit(int(subcommands.Execute(ctx)))
}
