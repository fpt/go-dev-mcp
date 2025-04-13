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

type GithubCmd struct {
	cdr *subcommands.Commander
}

func (*GithubCmd) Name() string     { return "github" }
func (*GithubCmd) Synopsis() string { return "Interact with GitHub." }
func (*GithubCmd) Usage() string {
	return `github [flags]:
  Interact with GitHub.
`
}

func (c *GithubCmd) SetFlags(f *flag.FlagSet) {
	cdr := subcommands.NewCommander(f, "")

	cdr.Register(cdr.CommandsCommand(), "help")
	cdr.Register(cdr.FlagsCommand(), "help")
	cdr.Register(cdr.HelpCommand(), "help")
	cdr.Register(&SearchCodeCmd{}, "searchcode")
	cdr.Register(&GetContentCmd{}, "getcontent")

	c.cdr = cdr
}

func (c *GithubCmd) Execute(ctx context.Context, f *flag.FlagSet, args ...any) subcommands.ExitStatus {
	return c.cdr.Execute(ctx, args...)
}

type SearchCodeCmd struct {
	repo string
}

func (*SearchCodeCmd) Name() string     { return "searchcode" }
func (*SearchCodeCmd) Synopsis() string { return "Search code on GitHub." }
func (*SearchCodeCmd) Usage() string {
	return `searchcode [flags]:
  Search for code on GitHub.
`
}

func (c *SearchCodeCmd) SetFlags(f *flag.FlagSet) {
	f.StringVar(&c.repo, "repo", "", "Repository to search in (owner/repo)")
}

func (c *SearchCodeCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...any) subcommands.ExitStatus {
	if f.NArg() < 1 {
		fmt.Println("Error: Missing search query.")
		return subcommands.ExitUsageError
	}

	query := f.Arg(0)
	if query == "" {
		fmt.Println("Error: Missing search query.")
		return subcommands.ExitUsageError
	}
	fmt.Println("Searching GitHub for:", query)

	gh, err := infra.NewGitHubClient()
	if err != nil {
		fmt.Println("Error:", err)
		return subcommands.ExitFailure
	}

	result, err := app.GitHubSearchCode(context.Background(), gh, query, &c.repo)
	if err != nil {
		fmt.Println("Error:", err)
		return subcommands.ExitFailure
	}
	fmt.Println(result)

	return subcommands.ExitSuccess
}

type GetContentCmd struct{}

func (*GetContentCmd) Name() string     { return "getcontent" }
func (*GetContentCmd) Synopsis() string { return "Get content of a file from GitHub." }
func (*GetContentCmd) Usage() string {
	return `getcontent [flags]:
  Get content of a file from GitHub.
`
}

func (*GetContentCmd) SetFlags(f *flag.FlagSet) {
}

func (*GetContentCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...any) subcommands.ExitStatus {
	if f.NArg() < 2 {
		fmt.Println("Error: Missing arguments.")
		fmt.Println("Usage: getcontent <owner/repo> <path>")
		return subcommands.ExitUsageError
	}

	ownerRepo := f.Arg(0)
	parts := strings.Split(ownerRepo, "/")
	if len(parts) != 2 {
		fmt.Println("Error: Invalid repo format, expected 'owner/repo'.")
		return subcommands.ExitUsageError
	}
	owner := parts[0]
	repo := parts[1]
	path := f.Arg(1)

	fmt.Println("Getting content from GitHub for:", owner, repo, path)

	gh, err := infra.NewGitHubClient()
	if err != nil {
		fmt.Println("Error:", err)
		return subcommands.ExitFailure
	}

	content, err := gh.GetContent(context.Background(), owner, repo, path)
	if err != nil {
		fmt.Println("Error:", err)
		return subcommands.ExitFailure
	}
	fmt.Println(content)

	return subcommands.ExitSuccess
}
