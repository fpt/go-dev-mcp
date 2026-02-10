package app

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Golden file conformance tests for dq-based docs.python.org HTML parsing.
// These tests ensure that changes to the dq package do not alter
// the output of parsePyDocPage or parsePyModIndex.
//
// To update golden files after intentional changes:
//
//	go test -v ./internal/app/ -run TestGolden_parsePyDoc -update

func TestGolden_parsePyDocPage(t *testing.T) {
	tests := []struct {
		name    string
		fixture string
		golden  string
	}{
		{"abc", "pydoc_doc_abc.html", "pydoc_doc_abc.golden"},
		{"json", "pydoc_doc_json.html", "pydoc_doc_json.golden"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc := loadHTMLFixture(t, tt.fixture)
			matched, got := parsePyDocPage(doc)
			require.True(t, matched, "parsePyDocPage should find section#module-*")

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

func TestGolden_parsePyModIndex(t *testing.T) {
	doc := loadHTMLFixture(t, "pydoc_modindex.html")
	matched, entries := parsePyModIndex(doc)
	require.True(t, matched, "parsePyModIndex should find table.modindextable")
	require.NotEmpty(t, entries, "should have parsed some module entries")

	// Convert entries to a formatted string for golden comparison
	got := formatPyModEntries(entries)

	goldenPath := "testdata/pydoc_modindex.golden"
	if *update {
		writeGolden(t, goldenPath, got)
		return
	}

	want := readGolden(t, goldenPath)
	assert.Equal(t, want, got, "output does not match golden file pydoc_modindex.golden")
}

func formatPyModEntries(entries []pyModEntry) string {
	var result string
	for _, e := range entries {
		result += "* " + e.Name + "\n"
		result += "\tURL: " + e.URL + "\n"
		if e.Description != "" {
			result += "\tDescription: " + e.Description + "\n"
		}
	}
	return result
}
