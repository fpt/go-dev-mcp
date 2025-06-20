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

const defaultLinesPerPage = 50

// SearchGoDocArgs represents arguments for Go documentation search
type SearchGoDocArgs struct {
	Query string `json:"query"`
}

// ReadGoDocArgs represents arguments for reading Go documentation
type ReadGoDocArgs struct {
	PackageURL string `json:"package_url"`
	Offset     int    `json:"offset,omitempty"`
	Limit      int    `json:"limit,omitempty"`
}

func searchGoDoc(
	ctx context.Context,
	request mcp.CallToolRequest,
	args SearchGoDocArgs,
) (*mcp.CallToolResult, error) {
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

	return mcp.NewToolResultText(
		fmt.Sprintf("Search results for '%s':\n%s", args.Query, results),
	), nil
}

func readGoDoc(
	ctx context.Context,
	request mcp.CallToolRequest,
	args ReadGoDocArgs,
) (*mcp.CallToolResult, error) {
	if args.PackageURL == "" {
		return mcp.NewToolResultError("Missing package URL"), nil
	}

	// Set default limit if not specified
	limit := args.Limit
	if limit == 0 {
		limit = defaultLinesPerPage
	}

	httpcli := infra.NewHttpClient()
	result, totalLines, hasMore, err := app.ReadGoDocPaged(
		httpcli,
		args.PackageURL,
		args.Offset,
		limit,
	)
	if err != nil {
		slog.ErrorContext(ctx, "readGoDoc", "error", err)
		return mcp.NewToolResultError(fmt.Sprintf("Error reading Go documentation: %v", err)), nil
	}

	// Calculate line range for display
	startLine := args.Offset + 1
	endLine := args.Offset + len(strings.Split(strings.TrimSpace(result), "\n"))
	if strings.TrimSpace(result) == "" {
		endLine = args.Offset
	}

	response := fmt.Sprintf("Documentation for '%s' (Lines %d-%d of %d):\n%s",
		args.PackageURL, startLine, endLine, totalLines, result)

	if hasMore {
		nextOffset := args.Offset + limit
		response += fmt.Sprintf("\n... (use offset=%d to see more)", nextOffset)
	}

	return mcp.NewToolResultText(response), nil
}
