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

type SearchPyDocArgs struct {
	Query string `json:"query"`
}

type ReadPyDocArgs struct {
	ModuleName string `json:"module_name"`
	Offset     int    `json:"offset,omitempty"`
	Limit      int    `json:"limit,omitempty"`
}

type SearchWithinPyDocArgs struct {
	ModuleName string `json:"module_name"`
	Keyword    string `json:"keyword"`
	MaxMatches int    `json:"max_matches,omitempty"`
}

func searchPyDoc(
	ctx context.Context,
	request mcp.CallToolRequest,
	args SearchPyDocArgs,
) (*mcp.CallToolResult, error) {
	if args.Query == "" {
		return mcp.NewToolResultError("Missing search query"), nil
	}

	httpcli := infra.NewHttpClient()
	results, err := app.SearchPyDoc(httpcli, args.Query)
	if err != nil {
		slog.ErrorContext(ctx, "searchPyDoc", "error", err)
		return mcp.NewToolResultError(
			fmt.Sprintf("Error searching Python documentation: %v", err),
		), nil
	}

	if len(results) == 0 {
		return mcp.NewToolResultText("No results found"), nil
	}

	return mcp.NewToolResultText(
		fmt.Sprintf("Search results for '%s':\n%s", args.Query, results),
	), nil
}

func readPyDoc(
	ctx context.Context,
	request mcp.CallToolRequest,
	args ReadPyDocArgs,
) (*mcp.CallToolResult, error) {
	if args.ModuleName == "" {
		return mcp.NewToolResultError("Missing module name"), nil
	}

	limit := args.Limit
	if limit == 0 {
		limit = app.DefaultLinesPerPage
	}

	httpcli := infra.NewHttpClient()
	result, totalLines, hasMore, err := app.ReadPyDocPaged(
		httpcli,
		args.ModuleName,
		args.Offset,
		limit,
	)
	if err != nil {
		slog.ErrorContext(ctx, "readPyDoc", "error", err)
		return mcp.NewToolResultError(
			fmt.Sprintf("Error reading Python documentation: %v", err),
		), nil
	}

	startLine := args.Offset + 1
	endLine := args.Offset + len(strings.Split(strings.TrimSpace(result), "\n"))
	if strings.TrimSpace(result) == "" {
		endLine = args.Offset
	}

	response := fmt.Sprintf("Documentation for '%s' (Lines %d-%d of %d):\n%s",
		args.ModuleName, startLine, endLine, totalLines, result)

	if hasMore {
		nextOffset := args.Offset + limit
		response += fmt.Sprintf("\n... (use offset=%d to see more)", nextOffset)
	}

	return mcp.NewToolResultText(response), nil
}

func searchWithinPyDoc(
	ctx context.Context,
	request mcp.CallToolRequest,
	args SearchWithinPyDocArgs,
) (*mcp.CallToolResult, error) {
	if args.ModuleName == "" {
		return mcp.NewToolResultError("Missing module name"), nil
	}
	if args.Keyword == "" {
		return mcp.NewToolResultError("Missing search keyword"), nil
	}

	maxMatches := args.MaxMatches
	if maxMatches == 0 {
		maxMatches = 10
	}

	httpcli := infra.NewHttpClient()
	result, err := app.SearchWithinPyDoc(httpcli, args.ModuleName, args.Keyword, maxMatches)
	if err != nil {
		slog.ErrorContext(ctx, "searchWithinPyDoc", "error", err)
		return mcp.NewToolResultError(
			fmt.Sprintf("Error searching Python documentation: %v", err),
		), nil
	}

	if len(result.Matches) == 0 {
		return mcp.NewToolResultText(
			fmt.Sprintf("No matches found for '%s' in module '%s'", args.Keyword, args.ModuleName),
		), nil
	}

	builder := strings.Builder{}
	builder.WriteString(
		fmt.Sprintf("Search results for '%s' in module '%s':\n\n", args.Keyword, args.ModuleName),
	)

	for _, match := range result.Matches {
		builder.WriteString(fmt.Sprintf("- Line %d\n```\n%s\n```\n", match.LineNo, match.Text))
	}

	if result.Truncated {
		builder.WriteString("... (additional matches truncated)\n")
	}

	return mcp.NewToolResultText(builder.String()), nil
}
