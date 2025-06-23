package tool

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/fpt/go-dev-mcp/internal/app"
	"github.com/fpt/go-dev-mcp/internal/infra"
	"github.com/mark3labs/mcp-go/mcp"
)

// ScanMarkdownArgs represents arguments for markdown file scanning
type ScanMarkdownArgs struct {
	Path string `json:"path"`
}

func scanMarkdown(
	ctx context.Context,
	request mcp.CallToolRequest,
	args ScanMarkdownArgs,
) (*mcp.CallToolResult, error) {
	if args.Path == "" {
		return mcp.NewToolResultError("path is required"), nil
	}

	fw := infra.NewFileWalker()
	results, err := app.ScanMarkdownFiles(ctx, fw, args.Path)
	if err != nil {
		slog.ErrorContext(ctx, "scanMarkdown", "error", err)
		return mcp.NewToolResultError(fmt.Sprintf("Error scanning markdown files: %v", err)), nil
	}

	output := app.FormatMarkdownScanResult(results)
	return mcp.NewToolResultText(output), nil
}
