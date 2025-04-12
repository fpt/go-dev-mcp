package htmlu

import (
	"strings"
	"testing"

	"golang.org/x/net/html"
)

func TestHasClass(t *testing.T) {
	tests := []struct {
		name     string
		nodeHTML string
		pattern  string
		want     bool
	}{
		{
			name:     "node with matching class",
			nodeHTML: `<div class="SearchSnippet">test</div>`,
			pattern:  "SearchSnippet",
			want:     true,
		},
		{
			name:     "node with multiple classes including match",
			nodeHTML: `<div class="SearchSnippet header primary">test</div>`,
			pattern:  "header",
			want:     true,
		},
		{
			name:     "node without matching class",
			nodeHTML: `<div class="OtherClass">test</div>`,
			pattern:  "SearchSnippet",
			want:     false,
		},
		{
			name:     "node with no class",
			nodeHTML: `<div>test</div>`,
			pattern:  "SearchSnippet",
			want:     false,
		},
		{
			name:     "wildcard pattern match",
			nodeHTML: `<div class="UnitDocHeader">test</div>`,
			pattern:  "Unit*",
			want:     true,
		},
		{
			name:     "wildcard pattern no match",
			nodeHTML: `<div class="OtherClass">test</div>`,
			pattern:  "Unit*",
			want:     false,
		},
		{
			name:     "question mark pattern match",
			nodeHTML: `<div class="SearchSnippet">test</div>`,
			pattern:  "Search?nippet",
			want:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := html.Parse(strings.NewReader(tt.nodeHTML))
			if err != nil {
				t.Fatalf("failed to parse HTML: %v", err)
			}

			// Find the div element
			var divNode *html.Node
			var findDiv func(*html.Node)
			findDiv = func(n *html.Node) {
				if n.Type == html.ElementNode && n.Data == "div" {
					divNode = n
					return
				}
				for c := n.FirstChild; c != nil; c = c.NextSibling {
					findDiv(c)
				}
			}
			findDiv(doc)

			if divNode == nil {
				t.Fatal("failed to find div element in test HTML")
			}

			if got := HasClass(divNode, tt.pattern); got != tt.want {
				t.Errorf("HasClass() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWalk(t *testing.T) {
	tests := []struct {
		name     string
		nodeHTML string
		want     []string
	}{
		{
			name:     "walk through nodes",
			nodeHTML: `<div class="SearchSnippet"><p>test</p></div>`,
			want:     []string{"div", "p"},
		},
		{
			name:     "nested nodes",
			nodeHTML: `<div><span><a href="#">link</a></span></div>`,
			want:     []string{"div", "span", "a"},
		},
		{
			name:     "flat list",
			nodeHTML: `<ul><li>Item 1</li><li>Item 2</li></ul>`,
			want:     []string{"ul", "li", "li"},
		},
		{
			name:     "nested list",
			nodeHTML: `<ul><li>Item 1</li><li><ul><li>Subitem 1</li></ul></li></ul>`,
			want:     []string{"ul", "li", "li", "ul", "li"},
		},
		{
			name:     "empty list",
			nodeHTML: `<ul></ul>`,
			want:     []string{"ul"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := html.Parse(strings.NewReader(tt.nodeHTML))
			if err != nil {
				t.Fatalf("failed to parse HTML: %v", err)
			}

			var got []string
			Walk(doc, func(n *html.Node) bool {
				if n.Type == html.ElementNode {
					got = append(got, n.Data)
				}
				return true
			})
			got = got[3:] // Skip [html head body]

			if len(got) != len(tt.want) {
				t.Errorf("Walk() got = %v, want %v", got, tt.want)
				return
			}

			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("Walk() got[%d] = %v, want %v", i, got[i], tt.want[i])
				}
			}
		})
	}
}
