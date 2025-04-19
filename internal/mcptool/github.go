package tool

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"fujlog.net/godev-mcp/internal/app"
	"fujlog.net/godev-mcp/internal/infra"
	"github.com/mark3labs/mcp-go/mcp"
)

func searchCodeGitHub(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	query, ok := request.Params.Arguments["query"].(string)
	if !ok || query == "" {
		return mcp.NewToolResultError("Missing search query"), nil
	}
	language, ok := request.Params.Arguments["language"].(string)
	if !ok || language == "" {
		return mcp.NewToolResultError("Missing language"), nil
	}
	ownerRepo, ok := request.Params.Arguments["repo"].(string)
	if ok && ownerRepo != "" {
		// Validate the repo format
		parts := strings.Split(ownerRepo, "/")
		if len(parts) != 2 {
			return mcp.NewToolResultError("Invalid repo format, expected 'owner/repo'"), nil
		}
	}

	gh, err := infra.NewGitHubClient()
	if err != nil {
		slog.ErrorContext(ctx, "searchCodeGitHub", "error", err)
		return mcp.NewToolResultError(fmt.Sprintf("Error creating GitHub client: %v", err)), nil
	}

	result, err := app.GitHubSearchCode(ctx, gh, query, &language, &ownerRepo)
	if err != nil {
		slog.ErrorContext(ctx, "searchCodeGitHub", "error", err)
		return mcp.NewToolResultError(fmt.Sprintf("Error searching code: %v", err)), nil
	}

	return mcp.NewToolResultText(result), nil
}

func getGitHubContent(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	ownerRepo, ok := request.Params.Arguments["repo"].(string)
	if !ok || ownerRepo == "" {
		return mcp.NewToolResultError("Missing repo"), nil
	}

	parts := strings.Split(ownerRepo, "/")
	if len(parts) != 2 {
		return mcp.NewToolResultError("Invalid repo format, expected 'owner/repo'"), nil
	}
	owner := parts[0]
	repo := parts[1]

	path, ok := request.Params.Arguments["path"].(string)
	if !ok || path == "" {
		return mcp.NewToolResultError("Missing path"), nil
	}

	gh, err := infra.NewGitHubClient()
	if err != nil {
		slog.ErrorContext(ctx, "getGitHubContent", "error", err)
		return mcp.NewToolResultError(fmt.Sprintf("Error creating GitHub client: %v", err)), nil
	}

	content, err := gh.GetContent(ctx, owner, repo, path)
	if err != nil {
		slog.ErrorContext(ctx, "getGitHubContent", "error", err)
		return mcp.NewToolResultError(fmt.Sprintf("Error getting content: %v", err)), nil
	}

	return mcp.NewToolResultText(content), nil
}

func getGitHubTree(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	ownerRepo, ok := request.Params.Arguments["repo"].(string)
	if !ok || ownerRepo == "" {
		return mcp.NewToolResultError("Missing repo"), nil
	}

	parts := strings.Split(ownerRepo, "/")
	if len(parts) != 2 {
		return mcp.NewToolResultError("Invalid repo format, expected 'owner/repo'"), nil
	}
	owner := parts[0]
	repo := parts[1]

	path, ok := request.Params.Arguments["path"].(string)
	if !ok {
		// Default to empty path (repository root)
		path = ""
	}

	gh, err := infra.NewGitHubClient()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Error creating GitHub client: %v", err)), nil
	}

	// Create a string builder for the tree output
	b := strings.Builder{}
	b.WriteString(fmt.Sprintf("%s/%s:%s\n", owner, repo, path))

	// Generate the tree using our PrintGitHubTree function
	if err := app.PrintGitHubTree(ctx, &b, gh, owner, repo, path); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Error generating tree: %v", err)), nil
	}

	return mcp.NewToolResultText(b.String()), nil
}
