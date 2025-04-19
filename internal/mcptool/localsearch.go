package tool

import (
	"context"
	"fmt"
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
		return mcp.NewToolResultError(fmt.Sprintf("Error getting local files: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Local files: %v", localFiles)), nil
}
