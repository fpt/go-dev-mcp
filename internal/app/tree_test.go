package app

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"fujlog.net/godev-mcp/internal/infra"
	"github.com/stretchr/testify/assert"
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
		if err := os.WriteFile(file, []byte("test"), 0o600); err != nil {
			t.Fatalf("Failed to create test file %s: %v", file, err)
		}
	}

	// Create a symbolic link if supported by the OS
	linkPath := filepath.Join(tempDir, "link.txt")
	targetPath := filepath.Join(tempDir, "file1.txt")
	// We'll ignore any error here since symlinks may not be supported on all platforms
	_ = os.Symlink(targetPath, linkPath)

	// Create dot files and directories to test ignore functionality
	dotFiles := []string{
		filepath.Join(tempDir, ".dotfile"),
		filepath.Join(tempDir, ".hidden"),
	}
	for _, file := range dotFiles {
		if err := os.WriteFile(file, []byte("hidden"), 0o600); err != nil {
			t.Fatalf("Failed to create dot file %s: %v", file, err)
		}
	}

	// Create a .git directory (should always be ignored)
	gitDir := filepath.Join(tempDir, ".git")
	if err := os.Mkdir(gitDir, 0o755); err != nil {
		t.Fatalf("Failed to create .git directory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(gitDir, "config"), []byte("git config"), 0o600); err != nil {
		t.Fatalf("Failed to create git config file: %v", err)
	}

	// Create a dot directory
	dotDir := filepath.Join(tempDir, ".dotdir")
	if err := os.Mkdir(dotDir, 0o755); err != nil {
		t.Fatalf("Failed to create dot directory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dotDir, "hidden.txt"), []byte("hidden"), 0o600); err != nil {
		t.Fatalf("Failed to create file in dot directory: %v", err)
	}

	// Test the printTree function (original test with ignoreDot=false)
	t.Run("print directory structure", func(t *testing.T) {
		b := &strings.Builder{}
		walker := infra.NewDirWalker()
		ctx := context.Background()
		err := PrintTree(ctx, b, walker, tempDir, false, 10)
		assert.NoError(t, err, "PrintTree should not fail")
		result := b.String()

		// Verify the result contains expected paths
		expectedPaths := []string{
			"file1.txt",
			"file2.txt",
			"subdir",
			"subfile.txt",
		}

		for _, path := range expectedPaths {
			assert.Contains(t, result, path, "printDirectory() result should contain %q", path)
		}

		// Check if we have the correct number of lines (at least one per entry plus possible link)
		lines := strings.Split(strings.TrimSpace(result), "\n")
		minExpectedLines := len(expectedPaths)
		assert.GreaterOrEqual(
			t,
			len(lines),
			minExpectedLines,
			"printDirectory() result has %d lines, expected at least %d",
			len(lines),
			minExpectedLines,
		)
	})

	// Test with a non-existent directory
	t.Run("non-existent directory", func(t *testing.T) {
		b := &strings.Builder{}
		nonExistentDir := filepath.Join(tempDir, "nonexistent")
		walker := infra.NewDirWalker()
		ctx := context.Background()
		err := PrintTree(ctx, b, walker, nonExistentDir, false, 10)
		// For a non-existent directory, we expect an error
		assert.Error(t, err, "PrintTree should fail for non-existent directory")
		assert.Contains(
			t,
			err.Error(),
			"no such file or directory",
			"Error should indicate file not found",
		)
	})

	// Test without ignoring dot files
	t.Run("print directory structure without ignoring dots", func(t *testing.T) {
		b := &strings.Builder{}
		walker := infra.NewDirWalker()
		ctx := context.Background()
		err := PrintTree(ctx, b, walker, tempDir, false, 10)
		assert.NoError(t, err, "PrintTree should not fail")
		result := b.String()

		// Verify the result contains expected paths including dot files
		expectedPaths := []string{
			"file1.txt",
			"file2.txt",
			"subdir",
			"subfile.txt",
			".dotfile",
			".hidden",
			".dotdir",
		}

		for _, path := range expectedPaths {
			assert.Contains(t, result, path, "printDirectory() result should contain %q", path)
		}

		// .git should always be ignored
		assert.NotContains(
			t,
			result,
			".git",
			"printDirectory() result should not contain .git directory",
		)
	})

	// Test with ignoring dot files
	t.Run("print directory structure with ignoring dots", func(t *testing.T) {
		b := &strings.Builder{}
		walker := infra.NewDirWalker()
		ctx := context.Background()
		err := PrintTree(ctx, b, walker, tempDir, true, 10)
		assert.NoError(t, err, "PrintTree should not fail")
		result := b.String()

		// Verify the result contains expected paths but not dot files
		expectedPaths := []string{
			"file1.txt",
			"file2.txt",
			"subdir",
			"subfile.txt",
		}

		for _, path := range expectedPaths {
			assert.Contains(t, result, path, "printDirectory() result should contain %q", path)
		}

		// Dot files and directories should be ignored
		ignoredPaths := []string{
			".dotfile",
			".hidden",
			".dotdir",
			".git",
		}

		for _, path := range ignoredPaths {
			assert.NotContains(
				t,
				result,
				path,
				"printDirectory() result should not contain %q when ignoring dots",
				path,
			)
		}
	})
}
