package tool

import (
	"context"
	"fmt"
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
		return mcp.NewToolResultError(fmt.Sprintf("Error creating GitHub client: %v", err)), nil
	}

	result, err := app.GitHubSearchCode(ctx, gh, query, &ownerRepo)
	if err != nil {
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
		return mcp.NewToolResultError(fmt.Sprintf("Error creating GitHub client: %v", err)), nil
	}

	content, err := gh.GetContent(ctx, owner, repo, path)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Error getting content: %v", err)), nil
	}

	return mcp.NewToolResultText(content), nil
}
