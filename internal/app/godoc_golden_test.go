package app

import (
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/html"
)

var update = flag.Bool("update", false, "update golden files")

// Golden file conformance tests for dq-based HTML parsing.
// These tests ensure that changes to the dq package do not alter
// the output of parseDocument, parseSearchResult, or parseReadme.
//
// To update golden files after intentional changes:
//
//	go test -v ./internal/app/ -run TestGolden -update

func TestGolden_parseDocument(t *testing.T) {
	tests := []struct {
		name    string
		fixture string
		golden  string
	}{
		{"net/http", "doc_net_http.html", "doc_net_http.golden"},
		{"database/sql", "doc_database_sql.html", "doc_database_sql.golden"},
		{"regexp", "doc_regexp.html", "doc_regexp.golden"},
		{"io/fs", "doc_io_fs.html", "doc_io_fs.golden"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc := loadHTMLFixture(t, tt.fixture)
			matched, got := parseDocument(doc)
			require.True(t, matched, "parseDocument should find div.Documentation")

			goldenPath := filepath.Join("testdata", tt.golden)
			if *update {
				writeGolden(t, goldenPath, got)
				return
			}

			want := readGolden(t, goldenPath)
			assert.Equal(t, want, got, "output does not match golden file %s", tt.golden)
		})
	}
}

func TestGolden_parseSearchResult(t *testing.T) {
	doc := loadHTMLFixture(t, "search_mcp_go.html")
	matched, got := parseSearchResult(doc)
	require.True(t, matched, "parseSearchResult should find div.SearchResults")

	goldenPath := filepath.Join("testdata", "search_mcp_go.golden")
	if *update {
		writeGolden(t, goldenPath, got)
		return
	}

	want := readGolden(t, goldenPath)
	assert.Equal(t, want, got, "output does not match golden file search_mcp_go.golden")
}

func TestGolden_parseReadme(t *testing.T) {
	doc := loadHTMLFixture(t, "readme_testify.html")
	matched, got := parseReadme(doc)
	require.True(t, matched, "parseReadme should find div.Overview-readmeContent")

	goldenPath := filepath.Join("testdata", "readme_testify.golden")
	if *update {
		writeGolden(t, goldenPath, got)
		return
	}

	want := readGolden(t, goldenPath)
	assert.Equal(t, want, got, "output does not match golden file readme_testify.golden")
}

func loadHTMLFixture(t *testing.T, name string) *html.Node {
	t.Helper()

	data, err := os.ReadFile(filepath.Join("testdata", name))
	require.NoError(t, err, "failed to read fixture %s", name)

	doc, err := html.Parse(strings.NewReader(string(data)))
	require.NoError(t, err, "failed to parse HTML fixture %s", name)

	return doc
}

func readGolden(t *testing.T, path string) string {
	t.Helper()

	data, err := os.ReadFile(path)
	require.NoError(t, err, "golden file not found: %s (run with -update to generate)", path)

	return string(data)
}

func writeGolden(t *testing.T, path, content string) {
	t.Helper()

	err := os.WriteFile(path, []byte(content), 0o644)
	require.NoError(t, err, "failed to write golden file %s", path)

	t.Logf("updated golden file: %s", path)
}
