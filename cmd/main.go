package main

import (
	"context"
	"flag"
	"os"

	"fujlog.net/godev-mcp/internal/subcmd"
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

	flag.Parse()
	ctx := context.Background()
	os.Exit(int(subcommands.Execute(ctx)))
}
