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
