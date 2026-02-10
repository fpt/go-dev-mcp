package dq_test

import (
	"strings"
	"testing"

	"github.com/fpt/go-dev-mcp/pkg/dq"
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

		// Attribute value matching: exact [attr=val]
		{
			name:     "attr exact match",
			html:     `<div role="note">Test</div>`,
			pattern:  "div[role=note]",
			expected: true,
		},
		{
			name:     "attr exact no match",
			html:     `<div role="alert">Test</div>`,
			pattern:  "div[role=note]",
			expected: false,
		},
		{
			name:     "attr exact any element",
			html:     `<span role="note">Test</span>`,
			pattern:  "[role=note]",
			expected: true,
		},

		// Attribute value matching: contains [attr*=val]
		{
			name:     "attr contains match",
			html:     `<a href="/3/library/json.html#module-json">json</a>`,
			pattern:  "a[href*=#module-]",
			expected: true,
		},
		{
			name:     "attr contains no match",
			html:     `<a href="/3/library/json.html">json</a>`,
			pattern:  "a[href*=#module-]",
			expected: false,
		},

		// Attribute value matching: prefix [attr^=val]
		{
			name:     "attr prefix match",
			html:     `<a href="/3/library/abc.html">abc</a>`,
			pattern:  "a[href^=/3/library/]",
			expected: true,
		},
		{
			name:     "attr prefix no match",
			html:     `<a href="/2/library/abc.html">abc</a>`,
			pattern:  "a[href^=/3/library/]",
			expected: false,
		},

		// Attribute value matching: suffix [attr$=val]
		{
			name:     "attr suffix match",
			html:     `<a href="/docs/abc.html">abc</a>`,
			pattern:  "a[href$=.html]",
			expected: true,
		},
		{
			name:     "attr suffix no match",
			html:     `<a href="/docs/abc.json">abc</a>`,
			pattern:  "a[href$=.html]",
			expected: false,
		},

		// Attribute existence still works
		{
			name:     "attr existence still works",
			html:     `<a href="/foo">link</a>`,
			pattern:  "a[href]",
			expected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			node := createHTMLNode(tc.html)
			result := matchSingle(node, tc.pattern)
			assert.Equal(
				t,
				tc.expected,
				result,
				"Pattern: %s should match: %v",
				tc.pattern,
				tc.expected,
			)
		})
	}
}

func TestInnerTextWithFilter(t *testing.T) {
	htmlSrc := `<html><body><div>Hello <span class="badge">v1.21</span> World</div></body></html>`
	doc, err := html.Parse(strings.NewReader(htmlSrc))
	if err != nil {
		t.Fatalf("failed to parse HTML: %v", err)
	}
	div := dq.FindOne(doc, "div")
	assert.NotNil(t, div)

	t.Run("default filter keeps all elements", func(t *testing.T) {
		text := dq.InnerTextWithFilter(div, true, dq.DefaultNodeFilter)
		assert.Contains(t, text, "v1.21")
		assert.Contains(t, text, "Hello")
		assert.Contains(t, text, "World")
	})

	t.Run("custom filter skips badge span", func(t *testing.T) {
		filter := func(n *html.Node) bool {
			if n.Data == "span" && dq.GetAttr(n, "class") == "badge" {
				return false
			}
			return true
		}
		text := dq.InnerTextWithFilter(div, true, filter)
		assert.NotContains(t, text, "v1.21")
		assert.Contains(t, text, "Hello")
		assert.Contains(t, text, "World")
	})

	t.Run("InnerText uses default filter", func(t *testing.T) {
		// InnerText and InnerTextWithFilter(DefaultNodeFilter) produce same result
		assert.Equal(t,
			dq.InnerText(div, true),
			dq.InnerTextWithFilter(div, true, dq.DefaultNodeFilter),
		)
	})
}

func TestGetAttr(t *testing.T) {
	htmlSrc := `<html><body><a href="/foo" data-id="42" role="note">link</a></body></html>`
	doc, err := html.Parse(strings.NewReader(htmlSrc))
	if err != nil {
		t.Fatalf("failed to parse HTML: %v", err)
	}
	a := dq.FindOne(doc, "a")
	assert.NotNil(t, a)

	assert.Equal(t, "/foo", dq.GetAttr(a, "href"))
	assert.Equal(t, "42", dq.GetAttr(a, "data-id"))
	assert.Equal(t, "note", dq.GetAttr(a, "role"))
	assert.Equal(t, "", dq.GetAttr(a, "missing"))

	// GetHref should return the same as GetAttr(n, "href")
	assert.Equal(t, dq.GetAttr(a, "href"), dq.GetHref(a))
}

func TestFindOne(t *testing.T) {
	htmlSrc := `
<html>
<body>
	<div class="outer">
		<div class="inner">
			<strong>v1.2.3</strong>
			<span>hello</span>
		</div>
		<p>paragraph</p>
	</div>
</body>
</html>`

	doc, err := html.Parse(strings.NewReader(htmlSrc))
	if err != nil {
		t.Fatalf("failed to parse HTML: %v", err)
	}

	t.Run("find by tag", func(t *testing.T) {
		n := dq.FindOne(doc, "strong")
		assert.NotNil(t, n)
		assert.Equal(t, "v1.2.3", dq.InnerText(n, false))
	})

	t.Run("find by class", func(t *testing.T) {
		n := dq.FindOne(doc, "div.inner")
		assert.NotNil(t, n)
		assert.Equal(t, "div", n.Data)
	})

	t.Run("not found", func(t *testing.T) {
		n := dq.FindOne(doc, "h1")
		assert.Nil(t, n)
	})

	t.Run("returns first match", func(t *testing.T) {
		n := dq.FindOne(doc, "div")
		assert.NotNil(t, n)
		// Should be the outer div (first in depth-first order)
		for _, attr := range n.Attr {
			if attr.Key == "class" {
				assert.Equal(t, "outer", attr.Val)
			}
		}
	})
}

func TestFindAll(t *testing.T) {
	htmlSrc := `
<html>
<body>
	<ul>
		<li>Item 1</li>
		<li>Item 2</li>
		<li>Item 3</li>
	</ul>
	<ol>
		<li>Ordered 1</li>
	</ol>
</body>
</html>`

	doc, err := html.Parse(strings.NewReader(htmlSrc))
	if err != nil {
		t.Fatalf("failed to parse HTML: %v", err)
	}

	t.Run("find all li", func(t *testing.T) {
		nodes := dq.FindAll(doc, "li")
		assert.Len(t, nodes, 4)
	})

	t.Run("find none", func(t *testing.T) {
		nodes := dq.FindAll(doc, "h1")
		assert.Empty(t, nodes)
	})
}

func TestFindDirectChild(t *testing.T) {
	htmlSrc := `
<html>
<body>
	<div>
		<span>direct</span>
		<p>
			<span>nested</span>
		</p>
	</div>
</body>
</html>`

	doc, err := html.Parse(strings.NewReader(htmlSrc))
	if err != nil {
		t.Fatalf("failed to parse HTML: %v", err)
	}

	div := dq.FindOne(doc, "div")
	assert.NotNil(t, div)

	t.Run("find direct child span", func(t *testing.T) {
		n := dq.FindDirectChild(div, "span")
		assert.NotNil(t, n)
		assert.Equal(t, "direct", dq.InnerText(n, false))
	})

	t.Run("find direct child p", func(t *testing.T) {
		n := dq.FindDirectChild(div, "p")
		assert.NotNil(t, n)
	})

	t.Run("no direct child h1", func(t *testing.T) {
		n := dq.FindDirectChild(div, "h1")
		assert.Nil(t, n)
	})
}

func TestFindWithAttrValue(t *testing.T) {
	htmlSrc := `
<html>
<body>
	<a href="/3/library/json.html#module-json">json</a>
	<a href="/3/library/os.html">os</a>
	<a href="/2/old/thing.html">old</a>
</body>
</html>`

	doc, err := html.Parse(strings.NewReader(htmlSrc))
	if err != nil {
		t.Fatalf("failed to parse HTML: %v", err)
	}

	t.Run("FindAll with attr contains", func(t *testing.T) {
		nodes := dq.FindAll(doc, "a[href*=#module-]")
		assert.Len(t, nodes, 1)
		assert.Equal(t, "json", dq.InnerText(nodes[0], false))
	})

	t.Run("FindAll with attr prefix", func(t *testing.T) {
		nodes := dq.FindAll(doc, "a[href^=/3/library/]")
		assert.Len(t, nodes, 2)
	})

	t.Run("FindOne with attr suffix", func(t *testing.T) {
		n := dq.FindOne(doc, "a[href$=.html]")
		assert.NotNil(t, n)
	})
}

func TestRecursiveNodeMatcher(t *testing.T) {
	// Test case 1: Simple nested structure ul > li
	t.Run("simple nested ul > li", func(t *testing.T) {
		htmlSrc := `
<html>
<body>
	<ul>
		<li>Item 1</li>
		<li>Item 2</li>
	</ul>
</body>
</html>`

		doc, err := html.Parse(strings.NewReader(htmlSrc))
		if err != nil {
			t.Fatalf("failed to parse HTML: %v", err)
		}

		var matchedItems []string
		matcher := dq.NewRecursiveNodeMatcher(
			"ul > li",
			func(n *html.Node) {
				text := dq.InnerText(n, true)
				matchedItems = append(matchedItems, text)
			},
			false, // not recursive
		)

		dq.Traverse(doc, []dq.Matcher{matcher})

		expected := []string{"Item 1", "Item 2"}
		assert.Equal(t, expected, matchedItems, "Should match both li elements")
	})

	// Test case 2: Nested lists with recursion ul > li > ul > li
	t.Run("recursive nested lists", func(t *testing.T) {
		htmlSrc := `
<html>
<body>
	<ul>
		<li>Item 1
			<ul>
				<li>Nested 1</li>
				<li>Nested 2</li>
			</ul>
		</li>
		<li>Item 2</li>
	</ul>
</body>
</html>`

		doc, err := html.Parse(strings.NewReader(htmlSrc))
		if err != nil {
			t.Fatalf("failed to parse HTML: %v", err)
		}

		var matchedItems []string
		matcher := dq.NewRecursiveNodeMatcher(
			"ul > li",
			func(n *html.Node) {
				text := strings.TrimSpace(dq.InnerText(n, false)) // Only direct text, not nested
				if text != "" {
					matchedItems = append(matchedItems, text)
				}
			},
			true, // recursive - should find nested ul > li patterns
		)

		dq.Traverse(doc, []dq.Matcher{matcher})

		// Should match: "Item 1", "Item 2" from outer ul, and "Nested 1", "Nested 2" from inner ul
		assert.Contains(t, matchedItems, "Item 2", "Should match outer li")
		assert.Contains(t, matchedItems, "Nested 1", "Should match nested li")
		assert.Contains(t, matchedItems, "Nested 2", "Should match nested li")
	})

	// Test case 3: Complex pattern ul > li > ol > li
	t.Run("complex nested pattern ul > li > ol > li", func(t *testing.T) {
		htmlSrc := `
<html>
<body>
	<ul>
		<li>List item
			<ol>
				<li>Ordered 1</li>
				<li>Ordered 2</li>
			</ol>
		</li>
	</ul>
</body>
</html>`

		doc, err := html.Parse(strings.NewReader(htmlSrc))
		if err != nil {
			t.Fatalf("failed to parse HTML: %v", err)
		}

		var matchedItems []string
		matcher := dq.NewRecursiveNodeMatcher(
			"ul > li > ol > li",
			func(n *html.Node) {
				text := dq.InnerText(n, true)
				matchedItems = append(matchedItems, text)
			},
			false, // not recursive
		)

		dq.Traverse(doc, []dq.Matcher{matcher})

		expected := []string{"Ordered 1", "Ordered 2"}
		assert.Equal(t, expected, matchedItems, "Should match ol > li elements inside ul > li")
	})

	// Test case 4: Pattern with classes and IDs
	t.Run("pattern with classes", func(t *testing.T) {
		htmlSrc := `
<html>
<body>
	<div class="container">
		<ul class="menu">
			<li class="item">Menu Item 1</li>
			<li class="item">Menu Item 2</li>
		</ul>
	</div>
</body>
</html>`

		doc, err := html.Parse(strings.NewReader(htmlSrc))
		if err != nil {
			t.Fatalf("failed to parse HTML: %v", err)
		}

		var matchedItems []string
		matcher := dq.NewRecursiveNodeMatcher(
			"div.container > ul.menu > li.item",
			func(n *html.Node) {
				text := dq.InnerText(n, true)
				matchedItems = append(matchedItems, text)
			},
			false,
		)

		dq.Traverse(doc, []dq.Matcher{matcher})

		expected := []string{"Menu Item 1", "Menu Item 2"}
		assert.Equal(
			t,
			expected,
			matchedItems,
			"Should match li.item elements with full class path",
		)
	})

	// Test case 5: Recursive with sub-matchers
	t.Run("recursive with sub-matchers", func(t *testing.T) {
		htmlSrc := `
<html>
<body>
	<ul>
		<li>
			<span>Item 1</span>
			<ul>
				<li><span>Nested 1</span></li>
			</ul>
		</li>
	</ul>
</body>
</html>`

		doc, err := html.Parse(strings.NewReader(htmlSrc))
		if err != nil {
			t.Fatalf("failed to parse HTML: %v", err)
		}

		var liItems []string
		var spanItems []string

		spanMatcher := dq.NewNodeMatcher(
			dq.NewMatchFunc("span"),
			func(n *html.Node) {
				text := dq.InnerText(n, true)
				spanItems = append(spanItems, text)
			},
		)

		matcher := dq.NewRecursiveNodeMatcher(
			"ul > li",
			func(n *html.Node) {
				// This handler won't get direct text since we have sub-matchers
				liItems = append(liItems, "li-matched")
			},
			true,        // recursive
			spanMatcher, // sub-matcher for spans
		)

		dq.Traverse(doc, []dq.Matcher{matcher})

		// The key thing is that we get both span items, even if li is matched more times
		assert.Contains(t, spanItems, "Item 1", "Should match span in outer li")
		assert.Contains(t, spanItems, "Nested 1", "Should match span in nested li")
		assert.Len(t, spanItems, 2, "Should match exactly 2 span elements")
	})
}
