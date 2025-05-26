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

// SearchLocalFilesArgs represents arguments for local file search
type SearchLocalFilesArgs struct {
	Path      string `json:"path"`
	Query     string `json:"query"`
	Extension string `json:"extension"`
}

func searchLocalFiles(ctx context.Context, request mcp.CallToolRequest, args SearchLocalFilesArgs) (*mcp.CallToolResult, error) {
	if args.Path == "" {
		return mcp.NewToolResultError("Missing search path"), nil
	}
	if args.Query == "" {
		return mcp.NewToolResultError("Missing search query"), nil
	}

	// Ensure the extension starts with a dot
	if args.Extension != "" && !strings.HasPrefix(args.Extension, ".") {
		args.Extension = "." + args.Extension
	}

	fw := infra.NewFileWalker()
	localFiles, err := app.SearchLocalFiles(ctx, fw, args.Path, args.Extension, args.Query)
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
