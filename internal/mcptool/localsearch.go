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

type searchLocalFilesArguments struct {
	Path      string `param:"path"`
	Query     string `param:"query"`
	Extension string `param:"extension"`
}

func searchLocalFiles(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args, err := decodeArguments[searchLocalFilesArguments](request.Params.Arguments)
	if err != nil {
		return mcp.NewToolResultError("Invalid parameters"), nil
	}

	if args.Path == "" {
		return mcp.NewToolResultError("Missing search path"), nil
	}
	if args.Query == "" {
		return mcp.NewToolResultError("Missing search query"), nil
	}

	extension := args.Extension
	if extension != "" && !strings.HasPrefix(extension, ".") {
		extension = "." + extension
	}

	fw := infra.NewFileWalker()
	localFiles, err := app.SearchLocalFiles(ctx, fw, args.Path, extension, args.Query)
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
