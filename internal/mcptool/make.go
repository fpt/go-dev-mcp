package tool

import (
	"context"
	"fmt"
	"log/slog"

	"fujlog.net/godev-mcp/internal/infra"
	"github.com/mark3labs/mcp-go/mcp"
)

func runMakeTarget(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	workdir, ok := request.Params.Arguments["work_dir"].(string)
	if !ok {
		return mcp.NewToolResultError("work_dir must be a string"), nil
	}
	target, ok := request.Params.Arguments["target"].(string)
	if !ok {
		return mcp.NewToolResultError("target must be a string"), nil
	}

	// Check if the Makefile exists in the current directory
	if !infra.IsFileExist(workdir, "Makefile") {
		slog.ErrorContext(ctx, "runMakeTarget", "error", fmt.Errorf("no Makefile found in directory: %s", workdir))
		return mcp.NewToolResultError(fmt.Sprintf("no Makefile found in directory: %s", workdir)), nil
	}

	stdout, stderr, exitCode, err := infra.Run(workdir, "make", target)
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
