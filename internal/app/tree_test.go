package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"fujlog.net/godev-mcp/internal/infra"
)

func TestPrintTree(t *testing.T) {
	// Create a temporary directory structure for testing
	tempDir, err := os.MkdirTemp("", "filesystem-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a subdirectory
	subDir := filepath.Join(tempDir, "subdir")
	if err := os.Mkdir(subDir, 0o755); err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}

	// Create some files
	files := []string{
		filepath.Join(tempDir, "file1.txt"),
		filepath.Join(tempDir, "file2.txt"),
		filepath.Join(subDir, "subfile.txt"),
	}

	for _, file := range files {
		if err := os.WriteFile(file, []byte("test"), 0o644); err != nil {
			t.Fatalf("Failed to create test file %s: %v", file, err)
		}
	}

	// Create a symbolic link if supported by the OS
	linkPath := filepath.Join(tempDir, "link.txt")
	targetPath := filepath.Join(tempDir, "file1.txt")
	// We'll ignore any error here since symlinks may not be supported on all platforms
	_ = os.Symlink(targetPath, linkPath)

	// Test the printTree function
	t.Run("print directory structure", func(t *testing.T) {
		b := &strings.Builder{}
		walker := infra.NewDirWalker()
		PrintTree(b, walker, tempDir, false)
		result := b.String()

		// Verify the result contains expected paths
		expectedPaths := []string{
			tempDir,
			"file1.txt",
			"file2.txt",
			"subdir",
			"subfile.txt",
		}

		for _, path := range expectedPaths {
			if !strings.Contains(result, path) {
				t.Errorf("printDirectory() result does not contain %q", path)
			}
		}

		// Check if we have the correct number of lines (at least one per entry plus possible link)
		lines := strings.Split(strings.TrimSpace(result), "\n")
		minExpectedLines := len(expectedPaths)
		if len(lines) < minExpectedLines {
			t.Errorf("printDirectory() result has %d lines, expected at least %d", len(lines), minExpectedLines)
		}
	})

	// Test with a non-existent directory
	t.Run("non-existent directory", func(t *testing.T) {
		b := &strings.Builder{}
		nonExistentDir := filepath.Join(tempDir, "nonexistent")
		walker := infra.NewDirWalker()
		PrintTree(b, walker, nonExistentDir, false)
		result := b.String()

		// Should only have the directory name itself as it doesn't exist
		if !strings.Contains(result, nonExistentDir) {
			t.Errorf("printDirectory() result does not contain directory name %q", nonExistentDir)
		}
	})
}
