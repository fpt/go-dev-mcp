package tool

import (
	"context"
	"fmt"
	"log/slog"

	"fujlog.net/godev-mcp/internal/app"
	"fujlog.net/godev-mcp/internal/infra"
	"github.com/mark3labs/mcp-go/mcp"
)

type searchGoDocArguments struct {
	Query string `param:"query"`
}

func searchGoDoc(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args, err := decodeArguments[searchGoDocArguments](request.Params.Arguments)
	if err != nil {
		return mcp.NewToolResultError("Invalid parameters"), nil
	}

	if args.Query == "" {
		return mcp.NewToolResultError("Missing search query"), nil
	}

	httpcli := infra.NewHttpClient()
	results, err := app.SearchGoDoc(httpcli, args.Query)
	if err != nil {
		slog.ErrorContext(ctx, "searchGoDoc", "error", err)
		return mcp.NewToolResultError(fmt.Sprintf("Error searching Go documentation: %v", err)), nil
	}

	if len(results) == 0 {
		return mcp.NewToolResultText("No results found"), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Search results for '%s':\n%s", args.Query, results)), nil
}

type readGoDocArguments struct {
	PackageURL string `param:"package_url"`
}

func readGoDoc(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args, err := decodeArguments[readGoDocArguments](request.Params.Arguments)
	if err != nil {
		return mcp.NewToolResultError("Invalid parameters"), nil
	}

	if args.PackageURL == "" {
		return mcp.NewToolResultError("Missing package URL"), nil
	}

	httpcli := infra.NewHttpClient()
	result, err := app.ReadGoDoc(httpcli, args.PackageURL)
	if err != nil {
		slog.ErrorContext(ctx, "readGoDoc", "error", err)
		return mcp.NewToolResultError(fmt.Sprintf("Error reading Go documentation: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Documentation for '%s':\n%s", args.PackageURL, result)), nil
}
