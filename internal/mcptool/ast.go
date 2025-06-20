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

// ExtractFunctionNamesArgs represents arguments for extracting function names
type ExtractFunctionNamesArgs struct {
	Directory string `json:"directory"`
}

func extractFunctionNames(
	ctx context.Context,
	request mcp.CallToolRequest,
	args ExtractFunctionNamesArgs,
) (*mcp.CallToolResult, error) {
	if args.Directory == "" {
		return mcp.NewToolResultError("Missing directory path"), nil
	}

	fw := infra.NewFileWalker()
	results, err := app.ExtractFunctionNames(ctx, fw, args.Directory)
	if err != nil {
		slog.ErrorContext(ctx, "extractFunctionNames", "error", err)
		return mcp.NewToolResultError(
			fmt.Sprintf("Error extracting function names: %v", err),
		), nil
	}

	builder := strings.Builder{}
	for _, result := range results {
		fileInfo := fmt.Sprintf("File: %s\n", result.Filename)
		for _, function := range result.Functions {
			fileInfo += fmt.Sprintf("- %s\n", function)
		}
		builder.WriteString(fileInfo)
	}

	if builder.Len() == 0 {
		return mcp.NewToolResultText("No Go functions found in the specified directory."), nil
	}

	return mcp.NewToolResultText(builder.String()), nil
}
