package tool

import (
	"context"
	"fmt"
	"strings"

	"fujlog.net/godev-mcp/internal/app"
	"fujlog.net/godev-mcp/internal/infra"
	"github.com/mark3labs/mcp-go/mcp"
)

func treeDir(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	rootDir, ok := request.Params.Arguments["root_dir"].(string)
	if !ok {
		return nil, fmt.Errorf("root_dir not found or not a string")
	}
	if rootDir == "" {
		return nil, fmt.Errorf("root_dir is empty")
	}

	b := strings.Builder{}
	b.WriteString(fmt.Sprintf("%s\n", rootDir))
	walker := infra.NewDirWalker()
	err := app.PrintTree(ctx, &b, walker, rootDir, false)
	if err != nil {
		return nil, fmt.Errorf("error printing tree: %v", err)
	}

	result := b.String()
	if result == "" {
		result = "No files found in the workspace directory."
	} else {
		result = fmt.Sprintf("Files in workspace directory:\n%s", result)
	}

	return mcp.NewToolResultText(result), nil
}
