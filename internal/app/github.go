package app

import (
	"context"
	"fmt"
	"strings"

	"fujlog.net/godev-mcp/internal/infra"
	"fujlog.net/godev-mcp/internal/repository"
	"github.com/pkg/errors"
)

func GitHubSearchCode(ctx context.Context, github repository.GitHubClient, query string, language, repo *string) (string, error) {
	// Perform the search
	opt := &repository.SearchCodeOption{
		Language: language,
		Repo:     repo,
	}
	result, err := github.SearchCode(ctx, query, opt)
	if err != nil {
		return "", errors.Wrap(err, "failed to search code")
	}

	if result.Total == 0 {
		return "", errors.New("no results found")
	}

	response := fmt.Sprintf("Total: %d\n", result.Total)
	for _, item := range result.Items {
		response += fmt.Sprintf("Path: %s, Repository: %s\n", item.Path, item.Repository)
		for _, fragment := range item.Fragments {
			response += "```\n"
			response += fmt.Sprintf("%s\n", fragment)
			response += "```\n"
		}
	}

	return response, nil
}

// PrintGitHubTree prints a tree representation of a GitHub repository path
// using the same formatting as PrintTree does for local directories.
func PrintGitHubTree(ctx context.Context, b *strings.Builder, client *infra.GitHubClient, owner, repo, path string) error {
	// Create a GitHub-specific walker
	walker := infra.NewGitHubDirWalker(ctx, client, owner, repo)

	// Use the existing PrintTree function with our GitHub-specific walker
	return PrintTree(ctx, b, walker, path, false)
}
