package tool

import (
	"context"

	"github.com/fpt/go-dev-mcp/internal/app"
	"github.com/mark3labs/mcp-go/mcp"
)

// ValidateGoCodeArgs represents the request parameters for Go code validation
type ValidateGoCodeArgs struct {
	Directory string `json:"directory"`
}

// validateGoCode handles the validate_go_code tool request
func validateGoCode(
	ctx context.Context,
	request mcp.CallToolRequest,
	args ValidateGoCodeArgs,
) (*mcp.CallToolResult, error) {
	if args.Directory == "" {
		return mcp.NewToolResultError("directory is required"), nil
	}
	report, err := app.ValidateGoCode(ctx, args.Directory)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	// Build text response
	result := report.Summary + "\n\n"

	// Add detailed results
	for _, validationResult := range report.Results {
		// Add check header
		statusEmoji := "✓"
		if validationResult.Status == "fail" {
			statusEmoji = "✗"
		} else if validationResult.Status == "error" {
			statusEmoji = "⚠"
		}

		result += statusEmoji + " " + validationResult.Check + "\n"
		result += "   " + validationResult.Summary + "\n"

		// Add output if present
		if validationResult.Output != "" {
			result += "   Output:\n" + indentText(validationResult.Output, "   ") + "\n"
		}
		result += "\n"
	}

	return mcp.NewToolResultText(result), nil
}

// indentText indents each line of text with the given prefix
func indentText(text, prefix string) string {
	if text == "" {
		return text
	}

	lines := []string{}
	for _, line := range splitLines(text) {
		lines = append(lines, prefix+line)
	}

	result := ""
	for i, line := range lines {
		if i > 0 {
			result += "\n"
		}
		result += line
	}
	return result
}

// splitLines splits text into lines, preserving empty lines
func splitLines(text string) []string {
	if text == "" {
		return []string{}
	}

	lines := []string{}
	current := ""

	for _, r := range text {
		if r == '\n' {
			lines = append(lines, current)
			current = ""
		} else {
			current += string(r)
		}
	}

	// Add the last line if it's not empty or if text doesn't end with newline
	if current != "" || len(text) > 0 && text[len(text)-1] != '\n' {
		lines = append(lines, current)
	}

	return lines
}
