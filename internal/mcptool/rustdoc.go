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

type SearchRustDocArgs struct {
	Query string `json:"query"`
}

type ReadRustDocArgs struct {
	CrateURL string `json:"crate_url"`
	Offset   int    `json:"offset,omitempty"`
	Limit    int    `json:"limit,omitempty"`
}

type SearchWithinRustDocArgs struct {
	CrateURL   string `json:"crate_url"`
	Keyword    string `json:"keyword"`
	MaxMatches int    `json:"max_matches,omitempty"`
}

func searchRustDoc(
	ctx context.Context,
	request mcp.CallToolRequest,
	args SearchRustDocArgs,
) (*mcp.CallToolResult, error) {
	if args.Query == "" {
		return mcp.NewToolResultError("Missing search query"), nil
	}

	httpcli := infra.NewHttpClient()
	results, err := app.SearchRustDoc(httpcli, args.Query)
	if err != nil {
		slog.ErrorContext(ctx, "searchRustDoc", "error", err)
		return mcp.NewToolResultError(
			fmt.Sprintf("Error searching Rust documentation: %v", err),
		), nil
	}

	if len(results) == 0 {
		return mcp.NewToolResultText("No results found"), nil
	}

	return mcp.NewToolResultText(
		fmt.Sprintf("Search results for '%s':\n%s", args.Query, results),
	), nil
}

func readRustDoc(
	ctx context.Context,
	request mcp.CallToolRequest,
	args ReadRustDocArgs,
) (*mcp.CallToolResult, error) {
	if args.CrateURL == "" {
		return mcp.NewToolResultError("Missing crate URL"), nil
	}

	limit := args.Limit
	if limit == 0 {
		limit = app.DefaultLinesPerPage
	}

	httpcli := infra.NewHttpClient()
	result, totalLines, hasMore, err := app.ReadRustDocPaged(
		httpcli,
		args.CrateURL,
		args.Offset,
		limit,
	)
	if err != nil {
		slog.ErrorContext(ctx, "readRustDoc", "error", err)
		return mcp.NewToolResultError(
			fmt.Sprintf("Error reading Rust documentation: %v", err),
		), nil
	}

	startLine := args.Offset + 1
	endLine := args.Offset + len(strings.Split(strings.TrimSpace(result), "\n"))
	if strings.TrimSpace(result) == "" {
		endLine = args.Offset
	}

	response := fmt.Sprintf("Documentation for '%s' (Lines %d-%d of %d):\n%s",
		args.CrateURL, startLine, endLine, totalLines, result)

	if hasMore {
		nextOffset := args.Offset + limit
		response += fmt.Sprintf("\n... (use offset=%d to see more)", nextOffset)
	}

	return mcp.NewToolResultText(response), nil
}

func searchWithinRustDoc(
	ctx context.Context,
	request mcp.CallToolRequest,
	args SearchWithinRustDocArgs,
) (*mcp.CallToolResult, error) {
	if args.CrateURL == "" {
		return mcp.NewToolResultError("Missing crate URL"), nil
	}
	if args.Keyword == "" {
		return mcp.NewToolResultError("Missing search keyword"), nil
	}

	maxMatches := args.MaxMatches
	if maxMatches == 0 {
		maxMatches = 10
	}

	httpcli := infra.NewHttpClient()
	result, err := app.SearchWithinRustDoc(httpcli, args.CrateURL, args.Keyword, maxMatches)
	if err != nil {
		slog.ErrorContext(ctx, "searchWithinRustDoc", "error", err)
		return mcp.NewToolResultError(
			fmt.Sprintf("Error searching Rust documentation: %v", err),
		), nil
	}

	if len(result.Matches) == 0 {
		return mcp.NewToolResultText(
			fmt.Sprintf("No matches found for '%s' in crate '%s'", args.Keyword, args.CrateURL),
		), nil
	}

	builder := strings.Builder{}
	builder.WriteString(
		fmt.Sprintf("Search results for '%s' in crate '%s':\n\n", args.Keyword, args.CrateURL),
	)

	for _, match := range result.Matches {
		builder.WriteString(fmt.Sprintf("- Line %d\n```\n%s\n```\n", match.LineNo, match.Text))
	}

	if result.Truncated {
		builder.WriteString("... (additional matches truncated)\n")
	}

	return mcp.NewToolResultText(builder.String()), nil
}
