package app

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fpt/go-dev-mcp/internal/repository"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

type MarkdownHeading struct {
	Title string
	Level int
	Line  int
}

type MarkdownFile struct {
	FileName string
	Headings []MarkdownHeading
}

// ScanMarkdownFiles scans a directory or single file for markdown files and extracts headings with line numbers
func ScanMarkdownFiles(
	ctx context.Context,
	fw repository.FileWalker,
	path string,
) ([]MarkdownFile, error) {
	var results []MarkdownFile

	// Check if path is a file or directory
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("failed to stat path %s: %w", path, err)
	}

	if info.IsDir() {
		// Directory case: use existing directory walking logic
		err := fw.Walk(ctx, func(filePath string) error {
			headings, err := extractMarkdownHeadings(filePath)
			if err != nil {
				// Skip files that can't be parsed instead of failing
				return nil
			}

			if len(headings) > 0 {
				results = append(results, MarkdownFile{
					FileName: filePath,
					Headings: headings,
				})
			}

			return nil
		}, path, ".md", true)
		if err != nil {
			return nil, err
		}
	} else {
		// File case: check if it's a markdown file and process it directly
		if !strings.HasSuffix(strings.ToLower(path), ".md") {
			return nil, fmt.Errorf("file %s is not a markdown file (.md)", path)
		}

		headings, err := extractMarkdownHeadings(path)
		if err != nil {
			return nil, fmt.Errorf("failed to extract headings from %s: %w", path, err)
		}

		if len(headings) > 0 {
			results = append(results, MarkdownFile{
				FileName: path,
				Headings: headings,
			})
		}
	}

	return results, nil
}

// extractMarkdownHeadings extracts headings from a markdown file with line numbers
func extractMarkdownHeadings(filePath string) ([]MarkdownHeading, error) {
	// Read the file
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	// Parse with goldmark
	md := goldmark.New()
	doc := md.Parser().Parse(text.NewReader(content))

	var headings []MarkdownHeading

	// Walk the AST to find headings
	err = ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}

		if heading, ok := n.(*ast.Heading); ok {
			// Extract heading text
			title := extractHeadingText(heading, content)

			// Calculate line number from byte position
			lines := heading.Lines()
			var lineNum int
			if lines.Len() > 0 {
				firstSegment := lines.At(0)
				lineNum = calculateLineNumber(content, firstSegment.Start)
			} else {
				lineNum = 1
			}

			headings = append(headings, MarkdownHeading{
				Title: title,
				Level: heading.Level,
				Line:  lineNum,
			})
		}

		return ast.WalkContinue, nil
	})
	if err != nil {
		return nil, err
	}

	return headings, nil
}

// extractHeadingText extracts the text content from a heading node
func extractHeadingText(heading *ast.Heading, source []byte) string {
	var buf bytes.Buffer

	for child := heading.FirstChild(); child != nil; child = child.NextSibling() {
		if textNode, ok := child.(*ast.Text); ok {
			segment := textNode.Segment
			buf.Write(segment.Value(source))
		}
	}

	return strings.TrimSpace(buf.String())
}

// calculateLineNumber converts a byte position to line number (1-based)
func calculateLineNumber(content []byte, pos int) int {
	if pos > len(content) {
		pos = len(content)
	}

	lineNum := 1
	for i := 0; i < pos; i++ {
		if content[i] == '\n' {
			lineNum++
		}
	}

	return lineNum
}

// FormatMarkdownScanResult formats the scan results as a readable string
func FormatMarkdownScanResult(files []MarkdownFile) string {
	if len(files) == 0 {
		return "No markdown files found."
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Found %d markdown file(s) with headings:\n\n", len(files)))

	for _, file := range files {
		// Convert to absolute path for clarity
		absPath, err := filepath.Abs(file.FileName)
		if err != nil {
			absPath = file.FileName // fallback to original path if conversion fails
		}
		result.WriteString(fmt.Sprintf("%s\n", absPath))

		for _, heading := range file.Headings {
			indent := strings.Repeat("  ", heading.Level-1)
			result.WriteString(fmt.Sprintf("%s%s %s (line %d)\n",
				indent,
				strings.Repeat("#", heading.Level),
				heading.Title,
				heading.Line,
			))
		}
		result.WriteString("\n")
	}

	return result.String()
}
