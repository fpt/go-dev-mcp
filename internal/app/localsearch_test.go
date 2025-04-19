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
		if err := os.WriteFile(filepath.Join(tempDir, filename), []byte(content), 0o644); err != nil {
			t.Fatalf("Failed to create test file %s: %v", filename, err)
		}
	}

	fw := infra.NewFileWalker()
	query := "test"
	results, err := SearchLocalFiles(context.Background(), fw, tempDir, ".txt", query)
	if err != nil {
		t.Fatalf("Error searching local files: %v", err)
	}

	if len(results) != len(files) {
		t.Fatalf("Expected %d results, got %d", len(files), len(results))
	}

	for _, result := range results {
		for _, match := range result.Matches {
			if !strings.Contains(match.Text, query) {
				t.Errorf("Expected query '%s' in file '%s', got '%s'", query, result.Filename, match.Text)
			}
		}
	}
}
