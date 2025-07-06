package tool

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/fpt/go-dev-mcp/internal/app"
	"github.com/fpt/go-dev-mcp/internal/infra"
	"github.com/mark3labs/mcp-go/mcp"
)

// SearchCodeGitHubArgs represents arguments for GitHub code search
type SearchCodeGitHubArgs struct {
	Query    string `json:"query"`
	Language string `json:"language"`
	Repo     string `json:"repo"`
}

// GitHubContentArgs represents arguments for GitHub content retrieval
type GitHubContentArgs struct {
	Repo   string `json:"repo"`
	Path   string `json:"path"`
	Offset int    `json:"offset,omitempty"`
	Limit  int    `json:"limit,omitempty"`
}

// GitHubTreeArgs represents arguments for GitHub tree display
type GitHubTreeArgs struct {
	Repo      string `json:"repo"`
	Path      string `json:"path"`
	IgnoreDot bool   `json:"ignore_dot"`
	MaxDepth  int    `json:"max_depth,omitempty"`
}

func searchCodeGitHub(
	ctx context.Context,
	request mcp.CallToolRequest,
	args SearchCodeGitHubArgs,
) (*mcp.CallToolResult, error) {
	if args.Query == "" {
		return mcp.NewToolResultError("Missing search query"), nil
	}
	if args.Language == "" {
		return mcp.NewToolResultError("Missing language"), nil
	}
	if args.Repo != "" {
		// Validate the repo format
		parts := strings.Split(args.Repo, "/")
		if len(parts) != 2 {
			return mcp.NewToolResultError("Invalid repo format, expected 'owner/repo'"), nil
		}
	}

	gh, err := infra.NewGitHubClient()
	if err != nil {
		slog.ErrorContext(ctx, "searchCodeGitHub", "error", err)
		return mcp.NewToolResultError(fmt.Sprintf("Error creating GitHub client: %v", err)), nil
	}

	result, err := app.GitHubSearchCode(ctx, gh, args.Query, &args.Language, &args.Repo)
	if err != nil {
		slog.ErrorContext(ctx, "searchCodeGitHub", "error", err)
		return mcp.NewToolResultError(fmt.Sprintf("Error searching code: %v", err)), nil
	}

	return mcp.NewToolResultText(result), nil
}

func getGitHubContent(
	ctx context.Context,
	request mcp.CallToolRequest,
	args GitHubContentArgs,
) (*mcp.CallToolResult, error) {
	if args.Repo == "" {
		return mcp.NewToolResultError("Missing repo"), nil
	}

	parts := strings.Split(args.Repo, "/")
	if len(parts) != 2 {
		return mcp.NewToolResultError("Invalid repo format, expected 'owner/repo'"), nil
	}
	owner := parts[0]
	repo := parts[1]

	if args.Path == "" {
		return mcp.NewToolResultError("Missing path"), nil
	}

	gh, err := infra.NewGitHubClient()
	if err != nil {
		slog.ErrorContext(ctx, "getGitHubContent", "error", err)
		return mcp.NewToolResultError(fmt.Sprintf("Error creating GitHub client: %v", err)), nil
	}

	content, err := gh.GetContent(ctx, owner, repo, args.Path)
	if err != nil {
		slog.ErrorContext(ctx, "getGitHubContent", "error", err)
		return mcp.NewToolResultError(fmt.Sprintf("Error getting content: %v", err)), nil
	}

	// Apply offset/limit if specified
	if args.Offset > 0 || args.Limit > 0 {
		content = paginateContent(content, args.Offset, args.Limit)
	}

	return mcp.NewToolResultText(content), nil
}

func getGitHubTree(
	ctx context.Context,
	request mcp.CallToolRequest,
	args GitHubTreeArgs,
) (*mcp.CallToolResult, error) {
	if args.Repo == "" {
		return mcp.NewToolResultError("Missing repo"), nil
	}

	parts := strings.Split(args.Repo, "/")
	if len(parts) != 2 {
		return mcp.NewToolResultError("Invalid repo format, expected 'owner/repo'"), nil
	}
	owner := parts[0]
	repo := parts[1]

	gh, err := infra.NewGitHubClient()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Error creating GitHub client: %v", err)), nil
	}

	// Create a string builder for the tree output
	b := strings.Builder{}
	b.WriteString(fmt.Sprintf("%s/%s:%s\n", owner, repo, args.Path))

	// Set default max depth if not specified
	maxDepth := args.MaxDepth
	if maxDepth == 0 {
		maxDepth = 3 // More conservative default for GitHub trees due to network overhead
	}

	// Generate the tree using our PrintGitHubTree function
	if err := app.PrintGitHubTree(ctx, &b, gh, owner, repo, args.Path, args.IgnoreDot, maxDepth); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Error generating tree: %v", err)), nil
	}

	return mcp.NewToolResultText(b.String()), nil
}

// paginateContent applies offset/limit to content by lines, similar to how read_godoc works
func paginateContent(content string, offset, limit int) string {
	lines := strings.Split(content, "\n")
	totalLines := len(lines)

	// Apply offset
	if offset >= totalLines {
		return fmt.Sprintf("(offset %d exceeds file length of %d lines)", offset, totalLines)
	}
	if offset > 0 {
		lines = lines[offset:]
	}

	// Apply limit (0 means no limit)
	if limit > 0 && limit < len(lines) {
		lines = lines[:limit]
		// Add truncation indicator
		lines = append(lines, fmt.Sprintf("... (showing lines %d-%d of %d total lines)",
			offset+1, offset+limit, totalLines))
	}

	return strings.Join(lines, "\n")
}
