package app

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Golden file conformance tests for dq-based docs.rs HTML parsing.
// These tests ensure that changes to the dq package do not alter
// the output of parseDocsRsDocument or parseDocsRsSearchResult.
//
// To update golden files after intentional changes:
//
//	go test -v ./internal/app/ -run TestGolden_parseDocsRs -update

func TestGolden_parseDocsRsDocument(t *testing.T) {
	tests := []struct {
		name    string
		fixture string
		golden  string
	}{
		{"serde", "rustdoc_doc_serde.html", "rustdoc_doc_serde.golden"},
		{"tokio", "rustdoc_doc_tokio.html", "rustdoc_doc_tokio.golden"},
		{"regex", "rustdoc_doc_regex.html", "rustdoc_doc_regex.golden"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc := loadHTMLFixture(t, tt.fixture)
			matched, got := parseDocsRsDocument(doc)
			require.True(t, matched, "parseDocsRsDocument should find section#main-content")

			goldenPath := "testdata/" + tt.golden
			if *update {
				writeGolden(t, goldenPath, got)
				return
			}

			want := readGolden(t, goldenPath)
			assert.Equal(t, want, got, "output does not match golden file %s", tt.golden)
		})
	}
}

func TestGolden_parseDocsRsSearchResult(t *testing.T) {
	doc := loadHTMLFixture(t, "rustdoc_search_serde.html")
	matched, got := parseDocsRsSearchResult(doc)
	require.True(t, matched, "parseDocsRsSearchResult should find div.recent-releases-container")

	goldenPath := "testdata/rustdoc_search_serde.golden"
	if *update {
		writeGolden(t, goldenPath, got)
		return
	}

	want := readGolden(t, goldenPath)
	assert.Equal(t, want, got, "output does not match golden file rustdoc_search_serde.golden")
}
