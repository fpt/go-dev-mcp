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

// ExtractFunctionNamesArgs represents arguments for extracting function names
type ExtractFunctionNamesArgs struct {
	Directory string `json:"directory"`
}

// ExtractCallGraphArgs represents arguments for extracting call graph
type ExtractCallGraphArgs struct {
	FilePath string `json:"file_path"`
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

func extractCallGraph(
	ctx context.Context,
	request mcp.CallToolRequest,
	args ExtractCallGraphArgs,
) (*mcp.CallToolResult, error) {
	if args.FilePath == "" {
		return mcp.NewToolResultError("Missing file path"), nil
	}

	result, err := app.ExtractCallGraph(args.FilePath)
	if err != nil {
		slog.ErrorContext(ctx, "extractCallGraph", "error", err)
		return mcp.NewToolResultError(
			fmt.Sprintf("Error extracting call graph: %v", err),
		), nil
	}

	builder := strings.Builder{}
	builder.WriteString(fmt.Sprintf("Call Graph for: %s\n\n", result.Filename))

	if len(result.CallGraph) == 0 {
		builder.WriteString("No exported functions found in the file.\n")
	} else {
		for _, entry := range result.CallGraph {
			builder.WriteString(fmt.Sprintf("Function: %s\n", entry.Function))
			if len(entry.Calls) == 0 {
				builder.WriteString("  - No function calls\n")
			} else {
				for _, call := range entry.Calls {
					if call.Package != "" {
						builder.WriteString(fmt.Sprintf("  - %s.%s\n", call.Package, call.Name))
					} else {
						builder.WriteString(fmt.Sprintf("  - %s (local)\n", call.Name))
					}
				}
			}
			builder.WriteString("\n")
		}
	}

	return mcp.NewToolResultText(builder.String()), nil
}
