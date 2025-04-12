package repository

import "context"

type SearchCodeResult struct {
	Total int
	Items []SearchCodeItem
}
type SearchCodeItem struct {
	Name       string
	Path       string
	Repository string
	Fragments  []string
}

type GitHubClient interface {
	SearchCode(ctx context.Context, query string) (SearchCodeResult, error)
	GetContent(ctx context.Context, owner, repo, path string) (string, error)
}
