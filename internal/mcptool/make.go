package tool

import (
	"context"
	"fmt"
	"log/slog"
	"regexp"
	"strings"

	"fujlog.net/godev-mcp/internal/infra"
	"github.com/mark3labs/mcp-go/mcp"
)

// High-priority error patterns (critical issues)
var highPriorityPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)\b(error|fail|failed|critical|fatal|panic|exception|abort)\b`),
	regexp.MustCompile(`(?i)(compilation failed|build failed|test failed)`),
	regexp.MustCompile(`(?i)make: \*\*\*`),
	regexp.MustCompile(`(?i)collect2: error`),
	regexp.MustCompile(`(?i)ld returned`),
	regexp.MustCompile(`(?i)undefined reference`),
}

// Medium-priority patterns (warnings)
var mediumPriorityPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)\b(warning|warn|deprecated|missing|timeout)\b`),
	regexp.MustCompile(`(?i)(not found|undefined)`),
}

// Success/info patterns for summary
var successPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)(built|compiled|success|completed|finished)`),
	regexp.MustCompile(`(?i)(test.*passed|ok.*test)`),
}

// RunMakeArgs represents arguments for running make targets
type RunMakeArgs struct {
	WorkDir string `json:"work_dir"`
	Target  string `json:"target"`
}

func runMakeTarget(
	ctx context.Context,
	request mcp.CallToolRequest,
	args RunMakeArgs,
) (*mcp.CallToolResult, error) {
	// Check if the Makefile exists in the current directory
	if !infra.IsFileExist(args.WorkDir, "Makefile") {
		slog.ErrorContext(
			ctx,
			"runMakeTarget",
			"error",
			fmt.Errorf("no Makefile found in directory: %s", args.WorkDir),
		)
		return mcp.NewToolResultError(
			fmt.Sprintf("no Makefile found in directory: %s", args.WorkDir),
		), nil
	}

	stdout, stderr, exitCode, err := infra.Run(args.WorkDir, "make", args.Target)
	if err != nil {
		slog.ErrorContext(ctx, "runMakeTarget", "error", err)
		return mcp.NewToolResultError(
			fmt.Sprintf(
				"❌ Command execution failed: %v\n%s",
				err,
				smartFilterOutput(stdout, stderr),
			),
		), nil
	}

	// Get smart filtered output
	filteredOutput := smartFilterOutput(stdout, stderr)

	if exitCode != 0 {
		// Command was executed but exited with a non-zero status
		return mcp.NewToolResultError(
			fmt.Sprintf("❌ Build failed (exit %d)\n%s", exitCode, filteredOutput),
		), nil
	}

	return mcp.NewToolResultText(
		fmt.Sprintf("✅ Build completed (exit %d)\n%s", exitCode, filteredOutput),
	), nil
}

// LineMatch represents a matched line with its priority
type LineMatch struct {
	LineNumber int
	Content    string
	Priority   int // 1 = high (errors), 2 = medium (warnings), 3 = success/info
}

// smartFilterOutput filters build output to focus on errors/warnings
func smartFilterOutput(stdout, stderr string) string {
	var allMatches []LineMatch

	// Process stderr (higher priority)
	if stderr != "" {
		stderrMatches := findImportantLines(stderr, 0) // 0 offset for stderr
		allMatches = append(allMatches, stderrMatches...)
	}

	// Process stdout
	if stdout != "" {
		stdoutOffset := 0
		if stderr != "" {
			stdoutOffset = len(strings.Split(stderr, "\n"))
		}
		stdoutMatches := findImportantLines(stdout, stdoutOffset)
		allMatches = append(allMatches, stdoutMatches...)
	}

	// If we found important lines, use filtered output
	if len(allMatches) > 0 {
		return formatFilteredOutput(allMatches, stdout, stderr)
	}

	// Fallback to truncated original output
	return formatTruncatedOutput(stdout, stderr)
}

// findImportantLines scans text for error/warning patterns
func findImportantLines(text string, lineOffset int) []LineMatch {
	lines := strings.Split(text, "\n")
	var matches []LineMatch

	for i, line := range lines {
		lineNum := lineOffset + i + 1
		priority := getLinePriority(line)

		if priority > 0 {
			matches = append(matches, LineMatch{
				LineNumber: lineNum,
				Content:    line,
				Priority:   priority,
			})
		}
	}

	return matches
}

// getLinePriority returns priority level (0 = not important, 1 = high, 2 = medium, 3 = info)
func getLinePriority(line string) int {
	// Check high-priority patterns first
	for _, pattern := range highPriorityPatterns {
		if pattern.MatchString(line) {
			return 1
		}
	}

	// Check medium-priority patterns
	for _, pattern := range mediumPriorityPatterns {
		if pattern.MatchString(line) {
			return 2
		}
	}

	// Check success patterns
	for _, pattern := range successPatterns {
		if pattern.MatchString(line) {
			return 3
		}
	}

	return 0
}

// formatFilteredOutput creates formatted output from filtered matches
func formatFilteredOutput(matches []LineMatch, stdout, stderr string) string {
	var result strings.Builder

	// Group by priority
	var highPriority, mediumPriority, infoPriority []LineMatch

	for _, match := range matches {
		switch match.Priority {
		case 1:
			highPriority = append(highPriority, match)
		case 2:
			mediumPriority = append(mediumPriority, match)
		case 3:
			infoPriority = append(infoPriority, match)
		}
	}

	// Show errors first
	if len(highPriority) > 0 {
		result.WriteString("🚨 Errors/Critical Issues:\n")
		for _, match := range highPriority {
			result.WriteString(fmt.Sprintf("  Line %d: %s\n", match.LineNumber, match.Content))
		}
		result.WriteString("\n")
	}

	// Show warnings
	if len(mediumPriority) > 0 {
		result.WriteString("⚠️  Warnings:\n")
		for _, match := range mediumPriority {
			result.WriteString(fmt.Sprintf("  Line %d: %s\n", match.LineNumber, match.Content))
		}
		result.WriteString("\n")
	}

	// Show success info (limited)
	if len(infoPriority) > 0 && len(highPriority) == 0 {
		result.WriteString("ℹ️  Summary:\n")
		// Only show first few success lines
		limit := 3
		if len(infoPriority) < limit {
			limit = len(infoPriority)
		}
		for i := 0; i < limit; i++ {
			match := infoPriority[i]
			result.WriteString(fmt.Sprintf("  %s\n", match.Content))
		}
	}

	// Add omitted line count
	totalLines := len(strings.Split(stdout, "\n")) + len(strings.Split(stderr, "\n"))
	shownLines := len(matches)
	if totalLines > shownLines {
		result.WriteString(fmt.Sprintf("... (%d lines omitted)\n", totalLines-shownLines))
	}

	return result.String()
}

// formatTruncatedOutput provides fallback truncated output
func formatTruncatedOutput(stdout, stderr string) string {
	var result strings.Builder

	// Show last 20 lines of stderr if present
	if stderr != "" {
		lines := strings.Split(stderr, "\n")
		start := 0
		if len(lines) > 20 {
			start = len(lines) - 20
			result.WriteString("... (stderr truncated)\n")
		}
		for i := start; i < len(lines); i++ {
			if strings.TrimSpace(lines[i]) != "" {
				result.WriteString(fmt.Sprintf("stderr: %s\n", lines[i]))
			}
		}
	}

	// Show last 15 lines of stdout if present and no stderr
	if stdout != "" && stderr == "" {
		lines := strings.Split(stdout, "\n")
		start := 0
		if len(lines) > 15 {
			start = len(lines) - 15
			result.WriteString("... (output truncated)\n")
		}
		for i := start; i < len(lines); i++ {
			if strings.TrimSpace(lines[i]) != "" {
				result.WriteString(fmt.Sprintf("%s\n", lines[i]))
			}
		}
	}

	return result.String()
}
