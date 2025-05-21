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

type searchGitHubCodeArguments struct {
	Query    string `param:"query"`
	Language string `param:"language"`
	Repo     string `param:"repo"`
}

func searchCodeGitHub(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args, err := decodeArguments[searchGitHubCodeArguments](request.Params.Arguments)
	if err != nil {
		return mcp.NewToolResultError("Invalid parameters"), nil
	}

	if args.Query == "" {
		return mcp.NewToolResultError("Missing search query"), nil
	}
	if args.Language == "" {
		return mcp.NewToolResultError("Missing language"), nil
	}

	ownerRepo := args.Repo
	if ownerRepo != "" {
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

	result, err := app.GitHubSearchCode(ctx, gh, args.Query, &args.Language, &ownerRepo)
	if err != nil {
		slog.ErrorContext(ctx, "searchCodeGitHub", "error", err)
		return mcp.NewToolResultError(fmt.Sprintf("Error searching code: %v", err)), nil
	}

	return mcp.NewToolResultText(result), nil
}

type getGitHubContentArguments struct {
	Repo string `param:"repo"`
	Path string `param:"path"`
}

func getGitHubContent(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args, err := decodeArguments[getGitHubContentArguments](request.Params.Arguments)
	if err != nil {
		return mcp.NewToolResultError("Invalid parameters"), nil
	}

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

	return mcp.NewToolResultText(content), nil
}

type GetGitHubTreeArguments struct {
	Repo string `param:"repo"`
	Path string `param:"path"`
}

func getGitHubTree(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args, err := decodeArguments[GetGitHubTreeArguments](request.Params.Arguments)
	if err != nil {
		return mcp.NewToolResultError("Invalid parameters"), nil
	}

	if args.Repo == "" {
		return mcp.NewToolResultError("Missing repo"), nil
	}

	parts := strings.Split(args.Repo, "/")
	if len(parts) != 2 {
		return mcp.NewToolResultError("Invalid repo format, expected 'owner/repo'"), nil
	}
	owner := parts[0]
	repo := parts[1]

	// Path has a default value in registration, so it should never be nil
	path := args.Path

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
