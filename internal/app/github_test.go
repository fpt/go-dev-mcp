package app

import (
	"context"
	"strings"
	"testing"

	"github.com/fpt/go-dev-mcp/internal/repository"
)

// MockGitHubClient implements the GitHubClient interface for testing
type MockGitHubClient struct {
	searchResult repository.SearchCodeResult
	searchError  error
}

func (m *MockGitHubClient) SearchCode(
	ctx context.Context,
	query string,
	opt *repository.SearchCodeOption,
) (repository.SearchCodeResult, error) {
	return m.searchResult, m.searchError
}

func (m *MockGitHubClient) GetContent(
	ctx context.Context,
	owner, repo, path string,
) (string, error) {
	return "", nil // Not used in this test
}

func TestGitHubSearchCodeCompactFormat(t *testing.T) {
	// Setup mock client with test data
	mockClient := &MockGitHubClient{
		searchResult: repository.SearchCodeResult{
			Total: 2,
			Items: []repository.SearchCodeItem{
				{
					Name:       "main.go",
					Path:       "cmd/main.go",
					Repository: "owner/repo",
					Fragments:  []string{"func main() {", "    fmt.Println(\"test\")"},
				},
				{
					Name:       "utils.go",
					Path:       "pkg/utils.go",
					Repository: "owner/repo",
					Fragments:  []string{"func testFunc() {"},
				},
			},
		},
	}

	// Call the function
	result, err := GitHubSearchCode(context.Background(), mockClient, "test", nil, nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify compact format
	lines := strings.Split(strings.TrimSpace(result), "\n")

	// Should start with total count
	if !strings.HasPrefix(lines[0], "Total: 2") {
		t.Errorf("Expected total count, got: %s", lines[0])
	}

	// Check compact format: repository/path (no verbose labels)
	expectedCompactLines := []string{
		"owner/repo/cmd/main.go",
		"owner/repo/pkg/utils.go",
	}

	foundCompactLines := 0
	for _, line := range lines {
		for _, expected := range expectedCompactLines {
			if line == expected {
				foundCompactLines++
				break
			}
		}
	}

	if foundCompactLines != len(expectedCompactLines) {
		t.Errorf("Expected %d compact format lines, found %d. Full result:\n%s",
			len(expectedCompactLines), foundCompactLines, result)
	}

	// Verify no verbose labels are present
	if strings.Contains(result, "Path:") || strings.Contains(result, "Repository:") {
		t.Error("Result should not contain verbose labels like 'Path:' or 'Repository:'")
	}

	// Verify no markdown code blocks
	if strings.Contains(result, "```") {
		t.Error("Result should not contain markdown code blocks")
	}

	// Verify fragments are indented with spaces
	if !strings.Contains(result, "  func main() {") {
		t.Error("Expected fragments to be indented with spaces")
	}
}
