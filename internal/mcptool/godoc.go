package tool

import (
	"context"
	"fmt"
	"log/slog"

	"fujlog.net/godev-mcp/internal/app"
	"fujlog.net/godev-mcp/internal/infra"
	"github.com/mark3labs/mcp-go/mcp"
)

func searchGoDoc(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	query, ok := request.Params.Arguments["query"].(string)
	if !ok || query == "" {
		return mcp.NewToolResultError("Missing search query"), nil
	}

	httpcli := infra.NewHttpClient()
	results, err := app.SearchGoDoc(httpcli, query)
	if err != nil {
		slog.ErrorContext(ctx, "searchGoDoc", "error", err)
		return mcp.NewToolResultError(fmt.Sprintf("Error searching Go documentation: %v", err)), nil
	}

	if len(results) == 0 {
		return mcp.NewToolResultText("No results found"), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Search results for '%s':\n%s", query, results)), nil
}

func readGoDoc(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	packageURL, ok := request.Params.Arguments["package_url"].(string)
	if !ok || packageURL == "" {
		return mcp.NewToolResultError("Missing package URL"), nil
	}

	httpcli := infra.NewHttpClient()
	result, err := app.ReadGoDoc(httpcli, packageURL)
	if err != nil {
		slog.ErrorContext(ctx, "readGoDoc", "error", err)
		return mcp.NewToolResultError(fmt.Sprintf("Error reading Go documentation: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Documentation for '%s':\n%s", packageURL, result)), nil
}
