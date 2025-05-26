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

// TreeDirArgs represents arguments for directory tree listing
type TreeDirArgs struct {
	RootDir string `json:"root_dir"`
}

func treeDir(ctx context.Context, request mcp.CallToolRequest, args TreeDirArgs) (*mcp.CallToolResult, error) {
	if args.RootDir == "" {
		return mcp.NewToolResultError("root_dir is required"), nil
	}

	b := strings.Builder{}
	b.WriteString(fmt.Sprintf("%s\n", args.RootDir))
	walker := infra.NewDirWalker()
	err := app.PrintTree(ctx, &b, walker, args.RootDir, false)
	if err != nil {
		slog.ErrorContext(ctx, "treeDir", "error", err)
		return mcp.NewToolResultError(fmt.Sprintf("Error printing tree: %v", err)), nil
	}

	result := b.String()
	if result == "" {
		result = "No files found in the workspace directory."
	} else {
		result = fmt.Sprintf("Files in workspace directory:\n%s", result)
	}

	return mcp.NewToolResultText(result), nil
}
