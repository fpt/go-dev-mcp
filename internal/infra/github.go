package infra

import (
	"context"
	"fmt"
	"strings"

	"fujlog.net/godev-mcp/internal/repository"
	"github.com/google/go-github/v71/github"
	"github.com/pkg/errors"
)

const ItemTypeDir = "dir"

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
func (c *GitHubClient) SearchCode(
	ctx context.Context, query string, opt *repository.SearchCodeOption,
) (repository.SearchCodeResult, error) {
	opts := &github.SearchOptions{Sort: "indexed", TextMatch: true}
	query = strings.TrimSpace(query)
	if opt != nil {
		if opt.Language != nil && *opt.Language != "" {
			query += fmt.Sprintf(" language:%s ", *opt.Language)
		}
		if opt.Repo != nil && *opt.Repo != "" {
			query += fmt.Sprintf(" repo:%s ", *opt.Repo)
		}
	}
	res, _, err := c.Search.Code(ctx, query, opts)
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

// GetContent retrieves the content of a file or directory in a GitHub repository using the GitHub API.
// If path points to a file, it returns the file content.
// If path points to a directory, it returns a formatted directory listing.
func (c *GitHubClient) GetContent(ctx context.Context, owner, repo, path string) (string, error) {
	fileContent, directoryContent, _, err := c.Repositories.GetContents(ctx, owner, repo, path, nil)
	if err != nil {
		return "", err
	}

	// Handle file content
	if fileContent != nil {
		if fileContent.Content == nil {
			return "", errors.New("content is empty")
		}

		contentStr, err := fileContent.GetContent()
		if err != nil {
			return "", errors.Wrap(err, "failed to get content")
		}

		return contentStr, nil
	}

	// Handle directory content
	if directoryContent != nil {
		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("Directory: %s\nContents:\n", path))

		for _, item := range directoryContent {
			itemType := *item.Type
			itemName := *item.Name

			// Add trailing slash for directories to make them easily identifiable
			if itemType == ItemTypeDir {
				itemName += "/"
			}

			sb.WriteString(fmt.Sprintf("- %s (%s", itemName, itemType))

			// Add size information for files
			if item.Size != nil {
				sb.WriteString(fmt.Sprintf(", %d bytes", *item.Size))
			}

			sb.WriteString(")\n")
		}

		return sb.String(), nil
	}

	return "", errors.New("no content found (both file and directory content are nil)")
}

// WalkContentsFunc is the function called for each file or directory visited by WalkContents.
// The path argument contains the path to the file or directory.
// The isDir argument is true if the path is a directory and false if it is a file.
// The depth argument indicates how deep in the tree the current item is.
// If the function returns an error, walking is stopped.
type WalkContentsFunc func(path string, isDir bool, depth int) error

// WalkContents walks the directory tree of a GitHub repository.
// It calls the provided WalkContentsFunc for each file or directory in the tree.
// The function is called first with the start path, and then with each file or directory found.
// If the start path is a file, the function is called only for that file.
// If the start path is a directory, the function is called for that directory and all files and directories in it.
func (c *GitHubClient) WalkContents(
	ctx context.Context,
	owner, repo, startPath string,
	fn WalkContentsFunc,
) error {
	return c.walkContentsRecursive(ctx, owner, repo, startPath, fn, 0)
}

// walkContentsRecursive is an internal helper function for WalkContents.
func (c *GitHubClient) walkContentsRecursive(
	ctx context.Context, owner, repo, path string, fn WalkContentsFunc, depth int,
) error {
	fileContent, directoryContent, _, err := c.Repositories.GetContents(ctx, owner, repo, path, nil)
	if err != nil {
		return err
	}

	// If it's a file, call the function with isDir = false
	if fileContent != nil {
		return fn(path, false, depth)
	}

	// If it's a directory, first call the function for the directory itself with isDir = true
	if err := fn(path, true, depth); err != nil {
		return err
	}

	// Then call the function for each item in the directory
	for _, item := range directoryContent {
		itemPath := *item.Path
		isDir := *item.Type == ItemTypeDir

		// Call the function for this item
		if err := fn(itemPath, isDir, depth+1); err != nil {
			return err
		}

		// If it's a directory, recursively walk it
		if isDir {
			if err := c.walkContentsRecursive(ctx, owner, repo, itemPath, fn, depth+1); err != nil {
				return err
			}
		}
	}

	return nil
}

// GitHubDirWalker implements the repository.DirWalker interface for GitHub repositories
type GitHubDirWalker struct {
	client *GitHubClient
	owner  string
	repo   string
}

// NewGitHubDirWalker creates a new GitHubDirWalker instance
func NewGitHubDirWalker(
	ctx context.Context,
	client *GitHubClient,
	owner, repo string,
) repository.DirWalker {
	return &GitHubDirWalker{
		client: client,
		owner:  owner,
		repo:   repo,
	}
}

// Walk implements the repository.DirWalker interface for GitHub repositories
func (w *GitHubDirWalker) Walk(
	ctx context.Context,
	function repository.WalkDirFunc,
	prefixFunc repository.WalkDirNextPrefixFunc,
	prefix, path string,
	ignoreDot bool,
	maxDepth int,
) error {
	return w.walkWithDepth(ctx, function, prefixFunc, prefix, path, ignoreDot, maxDepth, 0)
}

func (w *GitHubDirWalker) walkWithDepth(
	ctx context.Context,
	function repository.WalkDirFunc,
	prefixFunc repository.WalkDirNextPrefixFunc,
	prefix, path string,
	ignoreDot bool,
	maxDepth int,
	currentDepth int,
) error {
	// Get contents of the directory
	_, directoryContent, _, err := w.client.Repositories.GetContents(
		ctx,
		w.owner,
		w.repo,
		path,
		nil,
	)
	if err != nil {
		return errors.Wrap(err, "failed to get directory contents")
	}

	// Filter directory entries first
	filteredContent := make([]*github.RepositoryContent, 0)
	for _, item := range directoryContent {
		name := *item.Name

		// Always filter out .git directory
		if name == ".git" {
			continue
		}
		// Optionally filter out other dot files/directories
		if ignoreDot && strings.HasPrefix(name, ".") {
			continue
		}
		filteredContent = append(filteredContent, item)
	}

	// Process filtered directory entries
	for i, item := range filteredContent {
		isLastEntry := (i == len(filteredContent)-1)
		name := *item.Name
		isDir := *item.Type == ItemTypeDir

		// Call the function for this entry
		if err := function(name, prefix, isLastEntry); err != nil {
			return err
		}

		// If it's a directory, recursively walk it
		if isDir {
			// Check if we've reached the max depth
			if currentDepth >= maxDepth {
				continue
			}

			nextPrefix := prefixFunc(prefix, isLastEntry)
			subpath := path
			if subpath == "" {
				subpath = name
			} else {
				subpath = subpath + "/" + name
			}

			if err := w.walkWithDepth(
				ctx, function, prefixFunc, nextPrefix, subpath, ignoreDot, maxDepth, currentDepth+1,
			); err != nil {
				return err
			}
		}
	}

	return nil
}
