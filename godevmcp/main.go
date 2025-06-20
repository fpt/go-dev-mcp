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
	subcommands.Register(&subcmd.DatetimeCmd{}, "")
	subcommands.Register(&subcmd.GoDocCmd{}, "")
	subcommands.Register(&subcmd.GithubCmd{}, "")
	subcommands.Register(&subcmd.LocalSearchCmd{}, "")

	flag.Parse()
	ctx := context.Background()
	os.Exit(int(subcommands.Execute(ctx)))
}
