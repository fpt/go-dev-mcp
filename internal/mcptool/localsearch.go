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

func searchLocalFiles(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	path, ok := request.Params.Arguments["path"].(string)
	if !ok || path == "" {
		return mcp.NewToolResultError("Missing search path"), nil
	}
	query, ok := request.Params.Arguments["query"].(string)
	if !ok || query == "" {
		return mcp.NewToolResultError("Missing search query"), nil
	}
	extension, ok := request.Params.Arguments["extension"].(string)
	if !ok {
		extension = ""
	}
	if extension != "" && !strings.HasPrefix(extension, ".") {
		extension = "." + extension
	}

	fw := infra.NewFileWalker()
	localFiles, err := app.SearchLocalFiles(ctx, fw, path, extension, query)
	if err != nil {
		slog.ErrorContext(ctx, "searchLocalFiles", "error", err)
		return mcp.NewToolResultError(fmt.Sprintf("Error getting local files: %v", err)), nil
	}

	builder := strings.Builder{}
	for _, file := range localFiles {
		fileMatches := fmt.Sprintf("File: %s\n", file.Filename)
		for _, match := range file.Matches {
			fileMatches += fmt.Sprintf("- Line %d\n```\n%s\n```\n", match.LineNo, match.Text)
		}
		builder.WriteString(fileMatches)
	}

	return mcp.NewToolResultText(builder.String()), nil
}
