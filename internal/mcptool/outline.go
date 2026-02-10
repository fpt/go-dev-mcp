package tool

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/fpt/go-dev-mcp/internal/app"
	"github.com/fpt/go-dev-mcp/internal/infra"
	"github.com/mark3labs/mcp-go/mcp"
)

// OutlineGoPackageArgs represents arguments for the outline_go_package tool.
type OutlineGoPackageArgs struct {
	Directory        string `json:"directory"`
	SkipDependencies bool   `json:"skip_dependencies,omitempty"`
	SkipDeclarations bool   `json:"skip_declarations,omitempty"`
	SkipCallGraph    bool   `json:"skip_call_graph,omitempty"`
}

func outlineGoPackage(
	ctx context.Context,
	request mcp.CallToolRequest,
	args OutlineGoPackageArgs,
) (*mcp.CallToolResult, error) {
	if args.Directory == "" {
		return mcp.NewToolResultError("Missing directory path"), nil
	}

	fw := infra.NewFileWalker()
	output, err := app.OutlineGoPackage(ctx, fw, args.Directory, app.OutlineGoPackageOptions{
		SkipDependencies: args.SkipDependencies,
		SkipDeclarations: args.SkipDeclarations,
		SkipCallGraph:    args.SkipCallGraph,
	})
	if err != nil {
		slog.ErrorContext(ctx, "outlineGoPackage", "error", err)
		return mcp.NewToolResultError(
			fmt.Sprintf("Error generating package outline: %v", err),
		), nil
	}

	return mcp.NewToolResultText(output), nil
}
