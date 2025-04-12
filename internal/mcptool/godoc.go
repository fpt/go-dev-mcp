package tool

import (
	"context"
	"fmt"

	"fujlog.net/godev-mcp/internal/app"
	"github.com/mark3labs/mcp-go/mcp"
)

func searchGoDoc(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	query, ok := request.Params.Arguments["query"].(string)
	if !ok || query == "" {
		return mcp.NewToolResultError("Missing search query"), nil
	}

	results, err := app.SearchGoDoc(query)
	if err != nil {
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

	result, err := app.ReadGoDoc(packageURL)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Error reading Go documentation: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Documentation for '%s':\n%s", packageURL, result)), nil
}
