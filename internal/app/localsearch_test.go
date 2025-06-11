package app

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"fujlog.net/godev-mcp/internal/infra"
)

func TestSearchLocalFiles(t *testing.T) {
	// Create a temporary directory structure for testing
	tempDir, err := os.MkdirTemp("", "filesystem-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create some files with content
	files := map[string]string{
		"file1.txt": "This is a test file with some content.",
		"file2.txt": "Another test file with different content.",
		"file3.txt": "test",
	}

	for filename, content := range files {
		if err := os.WriteFile(filepath.Join(tempDir, filename), []byte(content), 0o600); err != nil {
			t.Fatalf("Failed to create test file %s: %v", filename, err)
		}
	}

	fw := infra.NewFileWalker()
	query := "test"
	results, err := SearchLocalFiles(context.Background(), fw, tempDir, ".txt", query, 10)
	if err != nil {
		t.Fatalf("Error searching local files: %v", err)
	}

	if len(results) != len(files) {
		t.Fatalf("Expected %d results, got %d", len(files), len(results))
	}

	for _, result := range results {
		for _, match := range result.Matches {
			if !strings.Contains(match.Text, query) {
				t.Errorf(
					"Expected query '%s' in file '%s', got '%s'",
					query,
					result.Filename,
					match.Text,
				)
			}
		}
	}
}

func TestSearchLocalFilesMatchLimit(t *testing.T) {
	// Create a temporary directory structure for testing
	tempDir, err := os.MkdirTemp("", "filesystem-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a file with many matches
	content := "test line 1\ntest line 2\ntest line 3\ntest line 4\ntest line 5\n" +
		"test line 6\ntest line 7\ntest line 8\ntest line 9\ntest line 10\ntest line 11\ntest line 12"
	filename := "many_matches.txt"
	if err := os.WriteFile(filepath.Join(tempDir, filename), []byte(content), 0o600); err != nil {
		t.Fatalf("Failed to create test file %s: %v", filename, err)
	}

	fw := infra.NewFileWalker()
	query := "test"
	maxMatches := 5

	results, err := SearchLocalFiles(context.Background(), fw, tempDir, ".txt", query, maxMatches)
	if err != nil {
		t.Fatalf("Error searching local files: %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}

	result := results[0]
	if len(result.Matches) != maxMatches {
		t.Errorf("Expected %d matches, got %d", maxMatches, len(result.Matches))
	}

	if !result.Truncated {
		t.Error("Expected result to be truncated")
	}

	// Test with no limit (0 should use default of 10)
	results2, err := SearchLocalFiles(context.Background(), fw, tempDir, ".txt", query, 15)
	if err != nil {
		t.Fatalf("Error searching local files: %v", err)
	}

	result2 := results2[0]
	if len(result2.Matches) != 12 { // Total number of lines with "test"
		t.Errorf("Expected 12 matches (all lines), got %d", len(result2.Matches))
	}

	if result2.Truncated {
		t.Error("Expected result to not be truncated when limit is higher than total matches")
	}
}
