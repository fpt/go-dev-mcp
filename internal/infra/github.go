package infra

import (
	"context"
	"fmt"
	"strings"

	"fujlog.net/godev-mcp/internal/repository"
	"github.com/google/go-github/v71/github"
	"github.com/pkg/errors"
)

type GitHubClient struct {
	*github.Client
}

func NewGitHubClient() (*GitHubClient, error) {
	stdout, _, exitCode, err := Run(".", "gh", "auth", "token")
	if err != nil {
		return nil, errors.Wrap(err, "failed to authenticate with GitHub")
	}
	if exitCode < 0 {
		return nil, errors.New("failed to run gh command")
	}
	if exitCode != 0 {
		return nil, errors.New("failed to authenticate with GitHub")
	}

	token := strings.TrimSpace(stdout)
	if token == "" {
		return nil, errors.New("no token found")
	}

	// Create a new GitHub client with the token
	client := github.NewClient(nil).WithAuthToken(token)
	return &GitHubClient{Client: client}, nil
}

// SearchCode searches for code in a GitHub repository using the GitHub API.
// https://github.com/google/go-github/blob/b98b707876c8b20b0e1dbbdffb7898a5fcc2169d/github/search.go#L62
// GitHub API docs: https://docs.github.com/rest/search/search#search-code
func (c *GitHubClient) SearchCode(ctx context.Context, query string) (repository.SearchCodeResult, error) {
	opts := &github.SearchOptions{Sort: "indexed", TextMatch: true}
	res, _, err := c.Client.Search.Code(ctx, fmt.Sprintf("%s language:go", query), opts)
	if err != nil {
		return repository.SearchCodeResult{}, err
	}

	if *res.Total == 0 {
		return repository.SearchCodeResult{}, errors.New("no results found")
	}

	var result repository.SearchCodeResult
	result.Total = *res.Total
	result.Items = make([]repository.SearchCodeItem, len(res.CodeResults))
	for i, item := range res.CodeResults {
		frags := make([]string, len(item.TextMatches))
		for j, frag := range item.TextMatches {
			frags[j] = *frag.Fragment
		}

		result.Items[i] = repository.SearchCodeItem{
			Name:       *item.Name,
			Path:       *item.Path,
			Repository: *item.Repository.FullName,
			Fragments:  frags,
		}
	}
	return result, nil
}

// GetContent retrieves the content of a file in a GitHub repository using the GitHub API.
func (c *GitHubClient) GetContent(ctx context.Context, owner, repo, path string) (string, error) {
	content, _, _, err := c.Client.Repositories.GetContents(ctx, owner, repo, path, nil)
	if err != nil {
		return "", err
	}
	if content == nil {
		return "", errors.New("no content found")
	}
	if content.Content == nil {
		return "", errors.New("content is empty")
	}

	contentStr, err := content.GetContent()
	if err != nil {
		return "", errors.Wrap(err, "failed to get content")
	}

	return contentStr, nil
}
