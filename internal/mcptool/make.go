package tool

import (
	"context"
	"fmt"
	"log/slog"

	"fujlog.net/godev-mcp/internal/infra"
	"github.com/mark3labs/mcp-go/mcp"
)

type makeToolArguments struct {
	WorkDir string `param:"work_dir"`
	Target  string `param:"target"`
}

func runMakeTarget(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args, err := decodeArguments[makeToolArguments](request.Params.Arguments)
	if err != nil {
		return mcp.NewToolResultError("Invalid parameters"), nil
	}

	// Check if the Makefile exists in the current directory
	if !infra.IsFileExist(args.WorkDir, "Makefile") {
		slog.ErrorContext(ctx, "runMakeTarget", "error", fmt.Errorf("no Makefile found in directory: %s", args.WorkDir))
		return mcp.NewToolResultError(fmt.Sprintf("no Makefile found in directory: %s", args.WorkDir)), nil
	}

	stdout, stderr, exitCode, err := infra.Run(args.WorkDir, "make", args.Target)
	if err != nil {
		slog.ErrorContext(ctx, "runMakeTarget", "error", err)
		return mcp.NewToolResultError(
			fmt.Sprintf("Command execution failed. Error: %+v\n%s", err, formatOutput(stdout, stderr)),
		), nil
	}

	if exitCode != 0 {
		// Command was executed but exited with a non-zero status
		// We return it as an error so that the caller can recognize it.
		return mcp.NewToolResultError(
			fmt.Sprintf("Command failed. Exit code: %d\n%s", exitCode, formatOutput(stdout, stderr)),
		), nil
	}

	return mcp.NewToolResultText(
		fmt.Sprintf("Command succeeded. Exit code: %d\n%s", exitCode, formatOutput(stdout, stderr)),
	), nil
}

func formatOutput(stdout, stderr string) string {
	result := ""

	// Append stderr first so that it is at the top
	if stderr != "" {
		result = result + fmt.Sprintf("stderr:\n```\n%s\n```\n", stderr)
	}

	if stdout != "" {
		result = result + fmt.Sprintf("stdout:\n```\n%s\n```\n", stdout)
	}

	return result
}
