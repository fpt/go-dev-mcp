package dq_test

import (
	"strings"
	"testing"

	"fujlog.net/godev-mcp/pkg/dq"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/html"
)

func TestTraverse(t *testing.T) {
	htmlSrc := `
<html>
<body>
	<div class="Documentation">
		<div>
			<h1>Test</h1>
		</div>
	</div>
</body>
</html>
`

	doc, err := html.Parse(strings.NewReader(htmlSrc))
	if err != nil {
		t.Fatalf("failed to parse HTML: %v", err)
	}

	var matchedDoc bool
	var matchedH1 bool
	var innerText string
	matcher := dq.NewNodeMatcher(
		dq.NewMatchFunc("div.Documentation"),
		func(n *html.Node) {
			assert.Equal(t, "div", n.Data)
			matchedDoc = true
		},
		dq.NewNodeMatcher(
			dq.NewMatchFunc("h1"),
			func(n *html.Node) {
				assert.Equal(t, "h1", n.Data)
				innerText = dq.InnerText(n, true)
				matchedH1 = true
			},
		),
	)

	dq.Traverse(doc, []dq.Matcher{matcher})
	assert.True(t, matchedDoc, "Expected a match but none was found.")
	assert.True(t, matchedH1, "Expected a match but none was found.")
	assert.Equal(t, "Test", innerText, "Expected inner text to be 'Test'.")
}

func TestMatchSingle(t *testing.T) {
	// Helper function to create HTML nodes for testing
	createHTMLNode := func(htmlStr string) *html.Node {
		doc, err := html.Parse(strings.NewReader(htmlStr))
		if err != nil {
			t.Fatalf("Failed to parse HTML: %v", err)
		}
		// Find the body element
		var body *html.Node
		var findBody func(*html.Node)
		findBody = func(n *html.Node) {
			if n.Type == html.ElementNode && n.Data == "body" {
				body = n
				return
			}
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				findBody(c)
			}
		}
		findBody(doc)

		// Return the first element in the body (our test element)
		for c := body.FirstChild; c != nil; c = c.NextSibling {
			if c.Type == html.ElementNode {
				return c
			}
		}
		t.Fatalf("Failed to find test element in HTML")
		return nil
	}

	// We'll use NewMatchFunc to test the same functionality as matchSingle
	matchSingle := func(n *html.Node, pattern string) bool {
		matcher := dq.NewMatchFunc(pattern)
		return matcher(n)
	}

	// Test cases for different selector patterns
	testCases := []struct {
		name     string
		html     string
		pattern  string
		expected bool
	}{
		// Basic element matching
		{
			name:     "match element by tag",
			html:     "<div>Test</div>",
			pattern:  "div",
			expected: true,
		},
		{
			name:     "non-matching element tag",
			html:     "<div>Test</div>",
			pattern:  "span",
			expected: false,
		},

		// Test comma selector (multiple elements)
		{
			name:     "match multiple elements with comma - first match",
			html:     "<div>Test</div>",
			pattern:  "div,span",
			expected: true,
		},
		{
			name:     "match multiple elements with comma - second match",
			html:     "<span>Test</span>",
			pattern:  "div,span",
			expected: true,
		},
		{
			name:     "match multiple elements with comma - no match",
			html:     "<p>Test</p>",
			pattern:  "div,span",
			expected: false,
		},

		// Test attribute selector
		{
			name:     "match element with attribute",
			html:     "<div id=\"test\">Test</div>",
			pattern:  "div[id]",
			expected: true,
		},
		{
			name:     "match any element with attribute",
			html:     "<span id=\"test\">Test</span>",
			pattern:  "[id]",
			expected: true,
		},
		{
			name:     "non-matching attribute",
			html:     "<div class=\"test\">Test</div>",
			pattern:  "div[id]",
			expected: false,
		},

		// Test ID selector
		{
			name:     "match element with ID",
			html:     "<div id=\"test\">Test</div>",
			pattern:  "div#test",
			expected: true,
		},
		{
			name:     "match any element with ID",
			html:     "<span id=\"test\">Test</span>",
			pattern:  "#test",
			expected: true,
		},
		{
			name:     "non-matching ID",
			html:     "<div id=\"other\">Test</div>",
			pattern:  "div#test",
			expected: false,
		},

		// Test class selector
		{
			name:     "match element with class",
			html:     "<div class=\"test\">Test</div>",
			pattern:  "div.test",
			expected: true,
		},
		{
			name:     "match any element with class",
			html:     "<span class=\"test\">Test</span>",
			pattern:  ".test",
			expected: true,
		},
		{
			name:     "match element with multiple classes",
			html:     "<div class=\"test other class\">Test</div>",
			pattern:  "div.test",
			expected: true,
		},
		{
			name:     "non-matching class",
			html:     "<div class=\"other\">Test</div>",
			pattern:  "div.test",
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			node := createHTMLNode(tc.html)
			result := matchSingle(node, tc.pattern)
			assert.Equal(t, tc.expected, result, "Pattern: %s should match: %v", tc.pattern, tc.expected)
		})
	}
}
