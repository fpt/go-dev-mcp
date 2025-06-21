package contentsearch

import (
	"strings"
	"testing"

	"github.com/fpt/go-dev-mcp/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestSearchInContent(t *testing.T) {
	tests := []struct {
		name       string
		content    string
		query      string
		maxMatches int
		expected   []model.SearchMatch
		truncated  bool
		wantError  bool
	}{
		{
			name:       "basic search with matches",
			content:    "line one\nline two\nline three",
			query:      "line",
			maxMatches: 10,
			expected: []model.SearchMatch{
				{LineNo: 1, Text: "line one"},
				{LineNo: 2, Text: "line two"},
				{LineNo: 3, Text: "line three"},
			},
			truncated: false,
		},
		{
			name:       "case sensitive search",
			content:    "Hello World\nhello world\nHELLO WORLD",
			query:      "hello",
			maxMatches: 10,
			expected: []model.SearchMatch{
				{LineNo: 2, Text: "hello world"},
			},
			truncated: false,
		},
		{
			name:       "no matches found",
			content:    "line one\nline two\nline three",
			query:      "xyz",
			maxMatches: 10,
			expected:   nil,
			truncated:  false,
		},
		{
			name:       "max matches limit with truncation",
			content:    "test line 1\ntest line 2\ntest line 3\ntest line 4",
			query:      "test",
			maxMatches: 2,
			expected: []model.SearchMatch{
				{LineNo: 1, Text: "test line 1"},
				{LineNo: 2, Text: "test line 2"},
			},
			truncated: true,
		},
		{
			name:       "single line match",
			content:    "single line with keyword",
			query:      "keyword",
			maxMatches: 5,
			expected: []model.SearchMatch{
				{LineNo: 1, Text: "single line with keyword"},
			},
			truncated: false,
		},
		{
			name:       "empty content",
			content:    "",
			query:      "anything",
			maxMatches: 10,
			expected:   nil,
			truncated:  false,
		},
		{
			name:       "empty query",
			content:    "line one\nline two",
			query:      "",
			maxMatches: 10,
			expected: []model.SearchMatch{
				{LineNo: 1, Text: "line one"},
				{LineNo: 2, Text: "line two"},
			},
			truncated: false,
		},
		{
			name:       "max matches is zero",
			content:    "test line 1\ntest line 2",
			query:      "test",
			maxMatches: 0,
			expected:   nil,
			truncated:  true,
		},
		{
			name:       "partial word match",
			content:    "testing\ntester\ntest",
			query:      "test",
			maxMatches: 10,
			expected: []model.SearchMatch{
				{LineNo: 1, Text: "testing"},
				{LineNo: 2, Text: "tester"},
				{LineNo: 3, Text: "test"},
			},
			truncated: false,
		},
		{
			name:       "special characters in content",
			content:    "line with @#$%\nline with spaces\nline with\ttabs",
			query:      "with",
			maxMatches: 10,
			expected: []model.SearchMatch{
				{LineNo: 1, Text: "line with @#$%"},
				{LineNo: 2, Text: "line with spaces"},
				{LineNo: 3, Text: "line with\ttabs"},
			},
			truncated: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.content)
			matches, truncated, err := SearchInContent(reader, tt.query, tt.maxMatches)

			if tt.wantError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.expected, matches)
			assert.Equal(t, tt.truncated, truncated)
		})
	}
}

func TestSearchInLine(t *testing.T) {
	tests := []struct {
		name     string
		line     string
		query    string
		expected bool
	}{
		{
			name:     "exact match",
			line:     "hello world",
			query:    "hello world",
			expected: true,
		},
		{
			name:     "partial match",
			line:     "hello world",
			query:    "world",
			expected: true,
		},
		{
			name:     "case sensitive no match",
			line:     "Hello World",
			query:    "hello",
			expected: false,
		},
		{
			name:     "case sensitive match",
			line:     "Hello World",
			query:    "Hello",
			expected: true,
		},
		{
			name:     "no match",
			line:     "hello world",
			query:    "xyz",
			expected: false,
		},
		{
			name:     "empty query matches everything",
			line:     "any line",
			query:    "",
			expected: true,
		},
		{
			name:     "empty line with non-empty query",
			line:     "",
			query:    "test",
			expected: false,
		},
		{
			name:     "empty line with empty query",
			line:     "",
			query:    "",
			expected: true,
		},
		{
			name:     "special characters",
			line:     "line with @#$% symbols",
			query:    "@#$%",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := searchInLine(tt.line, tt.query)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSearchInContent_LargeContent(t *testing.T) {
	// Test with a larger content to ensure performance and correctness
	var lines []string
	for i := 1; i <= 1000; i++ {
		if i%100 == 0 {
			lines = append(lines, "special line number")
		} else {
			lines = append(lines, "regular line number")
		}
	}
	content := strings.Join(lines, "\n")

	reader := strings.NewReader(content)
	matches, truncated, err := SearchInContent(reader, "special", 5)

	assert.NoError(t, err)
	assert.True(t, truncated, "Should be truncated with only 5 max matches")
	assert.Len(t, matches, 5, "Should return exactly 5 matches")

	// Verify the first few matches are correct (lines 100, 200, 300, 400, 500)
	assert.Equal(t, 100, matches[0].LineNo)
	assert.Contains(t, matches[0].Text, "special")
	assert.Equal(t, 200, matches[1].LineNo)
	assert.Contains(t, matches[1].Text, "special")
}

func TestSearchInContent_MultilineContent(t *testing.T) {
	content := `package main

import "fmt"

func main() {
	fmt.Println("Hello World")
	// This is a comment
	fmt.Printf("Testing %s", "format")
}
`

	reader := strings.NewReader(content)
	matches, truncated, err := SearchInContent(reader, "fmt", 10)

	assert.NoError(t, err)
	assert.False(t, truncated)
	assert.Len(t, matches, 3)

	expected := []model.SearchMatch{
		{LineNo: 3, Text: `import "fmt"`},
		{LineNo: 6, Text: `	fmt.Println("Hello World")`},
		{LineNo: 8, Text: `	fmt.Printf("Testing %s", "format")`},
	}

	assert.Equal(t, expected, matches)
}
