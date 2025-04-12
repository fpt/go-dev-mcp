package app

import (
	"context"
	"fmt"

	"fujlog.net/godev-mcp/internal/repository"
	"github.com/pkg/errors"
)

func GitHubSearchCode(ctx context.Context, github repository.GitHubClient, query string) (string, error) {
	// Perform the search
	result, err := github.SearchCode(ctx, query)
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
			response += "----\n"
			response += fmt.Sprintf("%s\n", fragment)
		}
		response += "----\n"
	}

	return response, nil
}
