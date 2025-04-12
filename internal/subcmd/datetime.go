package subcmd

import (
	"context"
	"flag"
	"fmt"

	"fujlog.net/godev-mcp/internal/app"
	"github.com/google/subcommands"
)

type DatetimeCmd struct {
}

func (*DatetimeCmd) Name() string     { return "datetime" }
func (*DatetimeCmd) Synopsis() string { return "List current date and time." }
func (*DatetimeCmd) Usage() string {
	return `datetime [flags]:
  Display the current date and time.
`
}

func (p *DatetimeCmd) SetFlags(f *flag.FlagSet) {
}

func (p *DatetimeCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...any) subcommands.ExitStatus {
	currentTime, err := app.CurrentDatetime()
	if err != nil {
		fmt.Println("Error:", err)
		return subcommands.ExitFailure
	}
	fmt.Println(currentTime)

	return subcommands.ExitSuccess
}
