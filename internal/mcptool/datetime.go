package tool

import (
	"context"
	"fmt"

	"fujlog.net/godev-mcp/internal/app"
	"github.com/mark3labs/mcp-go/mcp"
)

func getCurrentDateTime(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	currentTime, err := app.CurrentDatetime()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Error getting current date and time: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Current date and time: %s", currentTime)), nil
}
