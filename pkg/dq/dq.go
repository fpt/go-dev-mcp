// Document Query - A simple HTML Document query library
package dq

import (
	"fmt"
	"path/filepath"
	"strings"

	"golang.org/x/net/html"
)

// Node Matcher
// Traverse the HTML node tree and match nodes based on the provided matcher.
// The structure of matcher is preserved in the tree.
// Traverse will visit nodes in the tree only once.
//
// This pkg doesn't provide a full CSS selector or jQuery selector.
//
// Selector
// ,: match multiple elements e.g. "div,span"
// []: match attribute e.g. "div[id]"
// #: match id e.g. "div#id"
// .: match class e.g. "div.class"
// >: match direct child e.g. ">span"
//
// RecursiveNodeMatcher extends the basic matching with:
// - Multi-step patterns: "ul > li > ol > li" - matches nested sequences
// - Recursive matching: can restart pattern after completion for deeply nested structures
// - Sibling matching: continues to match additional siblings at each level

type Matcher interface {
	Match(n *html.Node) bool
	Handler(n *html.Node)
	SubMatchers() []Matcher
}

type MatchFunc func(*html.Node) bool

type MatchHandlerFunc func(*html.Node)

type NodeMatcher struct {
	matchFunc   MatchFunc
	handler     MatchHandlerFunc
	subMatchers []Matcher
}

func NewNodeMatcher(
	matchFunc MatchFunc,
	handler MatchHandlerFunc,
	children ...Matcher,
) *NodeMatcher {
	return &NodeMatcher{
		matchFunc:   matchFunc,
		handler:     handler,
		subMatchers: children,
	}
}

func (m *NodeMatcher) Match(n *html.Node) bool {
	if m.matchFunc != nil {
		return m.matchFunc(n)
	}
	return false
}

func (m *NodeMatcher) Handler(n *html.Node) {
	if m.handler != nil {
		m.handler(n)
	}
}

func (m *NodeMatcher) SubMatchers() []Matcher {
	return m.subMatchers
}

func Traverse(n *html.Node, ms []Matcher) {
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode {
			var subMatchers []Matcher

			for _, m := range ms {
				if m.Match(c) {
					m.Handler(c)

					subMatchers = mergeUniqueMatchers(subMatchers, m.SubMatchers()...)
				} else {
					subMatchers = mergeUniqueMatchers(subMatchers, m)
				}
			}

			if len(subMatchers) == 0 {
				continue
			}

			// Traverse the child nodes of the current node
			Traverse(c, subMatchers)
		}
	}
}

func mergeUniqueMatchers(ma []Matcher, mb ...Matcher) []Matcher {
	seen := make(map[Matcher]bool)
	for _, m := range ma {
		seen[m] = true
	}
	for _, m := range mb {
		if !seen[m] {
			ma = append(ma, m)
			seen[m] = true
		}
	}
	return ma
}

func NewMatchFunc(pattern string) MatchFunc {
	pats := strings.Split(pattern, ",")

	return func(n *html.Node) bool {
		for _, pat := range pats {
			if matchSingle(n, strings.TrimSpace(pat)) {
				return true
			}
		}
		return false
	}
}

//nolint:nestif // TBD
func matchSingle(n *html.Node, pattern string) bool {
	if n.Type == html.ElementNode {
		if strings.Contains(pattern, "[") {
			ss := strings.Split(pattern, "[")
			element, attr := ss[0], ss[1]
			if (element == "" || n.Data == element) && hasAttr(n, attr[:len(attr)-1]) {
				return true
			}
		} else if strings.Contains(pattern, "#") {
			ss := strings.Split(pattern, "#")
			element, id := ss[0], ss[1]
			if (element == "" || n.Data == element) && hasID(n, id) {
				return true
			}
		} else if strings.Contains(pattern, ".") {
			ss := strings.Split(pattern, ".")
			element, class := ss[0], ss[1]
			if (element == "" || n.Data == element) && hasClass(n, class) {
				return true
			}
		} else {
			if n.Data == pattern {
				return true
			}
		}
	}
	return false
}

func hasID(n *html.Node, pattern string) bool {
	for _, attr := range n.Attr {
		if attr.Key == "id" {
			match, err := filepath.Match(pattern, attr.Val)
			if err == nil && match {
				return true
			}
			break
		}
	}
	return false
}

func hasClass(n *html.Node, pattern string) bool {
	for _, attr := range n.Attr {
		if attr.Key == "class" {
			classes := strings.Fields(attr.Val)
			for _, class := range classes { // Changed 'for class := range classes' to 'for _, class := range classes'
				match, err := filepath.Match(pattern, class)
				if err == nil && match {
					return true
				}
			}
			break
		}
	}
	return false
}

func hasAttr(n *html.Node, pattern string) bool {
	for _, attr := range n.Attr {
		match, err := filepath.Match(pattern, attr.Key)
		if err == nil && match {
			return true
		}
	}
	return false
}

func HasChild(n *html.Node, pattern string) bool {
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode && matchSingle(c, pattern) {
			return true
		}
	}
	return false
}

func GetHref(n *html.Node) string {
	for _, attr := range n.Attr {
		if attr.Key == "href" {
			return strings.TrimSpace(attr.Val)
		}
	}
	return ""
}

func InnerText(n *html.Node, recurse bool) string {
	var text string
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode && c.Data == "a" && hasAttr(c, "aria-label") {
			continue
		}

		if c.Type == html.ElementNode && c.Data == "span" &&
			hasClass(c, "Documentation-sinceVersion") {
			continue
		}

		if c.Type == html.TextNode {
			text += strings.TrimSpace(c.Data)
		}

		if recurse && c.Type == html.ElementNode {
			text += fmt.Sprintf(" %s ", InnerText(c, recurse))
		}
	}
	return strings.TrimSpace(text)
}

func RawInnerText(n *html.Node, recurse bool) string {
	if n.Type == html.TextNode {
		return n.Data
	}

	var text string
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode && c.Data == "a" && hasAttr(c, "aria-label") {
			continue
		}

		text += RawInnerText(c, recurse)
	}
	return text
}

// RecursiveNodeMatcher handles recursive pattern matching for nested structures
// like "ul > li > ol > li" where the pattern can repeat at different nesting levels
type RecursiveNodeMatcher struct {
	patterns    []string         // Sequence of patterns to match (e.g., ["ul", "li", "ol", "li"])
	currentStep int              // Current step in the pattern sequence
	handler     MatchHandlerFunc // Handler to call when pattern completes
	recursive   bool             // Whether to restart pattern after completion
	subMatchers []Matcher        // Additional sub-matchers to apply after pattern completion
	isRoot      bool             // Whether this is the root matcher (for sibling matching)
}

// NewRecursiveNodeMatcher creates a new recursive matcher
// pattern: space-separated pattern like "ul > li > ol > li" or "ul li ol li"
// handler: function to call when the full pattern matches
// recursive: if true, pattern restarts after completion for nested structures
// children: additional matchers to apply after pattern completion
func NewRecursiveNodeMatcher(
	pattern string,
	handler MatchHandlerFunc,
	recursive bool,
	children ...Matcher,
) *RecursiveNodeMatcher {
	// Parse pattern - handle both ">" and space-separated formats
	pattern = strings.ReplaceAll(pattern, ">", " ")
	patterns := strings.Fields(pattern)

	return &RecursiveNodeMatcher{
		patterns:    patterns,
		currentStep: 0,
		handler:     handler,
		recursive:   recursive,
		subMatchers: children,
		isRoot:      true,
	}
}

func (m *RecursiveNodeMatcher) Match(n *html.Node) bool {
	if m.currentStep >= len(m.patterns) {
		return false
	}

	currentPattern := strings.TrimSpace(m.patterns[m.currentStep])
	matcher := NewMatchFunc(currentPattern)
	result := matcher(n)
	return result
}

func (m *RecursiveNodeMatcher) Handler(n *html.Node) {
	// If we've completed the pattern, call the handler
	if m.currentStep+1 >= len(m.patterns) {
		if m.handler != nil {
			m.handler(n)
		}
	}
}

func (m *RecursiveNodeMatcher) SubMatchers() []Matcher {
	var matchers []Matcher

	// If we haven't completed the pattern, continue with next step
	if m.currentStep+1 < len(m.patterns) {
		// Create a new matcher for the next step
		nextMatcher := &RecursiveNodeMatcher{
			patterns:    m.patterns,
			currentStep: m.currentStep + 1,
			handler:     m.handler,
			recursive:   m.recursive,
			subMatchers: m.subMatchers,
			isRoot:      false, // Child matchers are not root
		}
		matchers = append(matchers, nextMatcher)

		// If this is the root matcher, also include itself for sibling matching
		if m.isRoot {
			rootMatcher := &RecursiveNodeMatcher{
				patterns:    m.patterns,
				currentStep: 0,
				handler:     m.handler,
				recursive:   m.recursive,
				subMatchers: m.subMatchers,
				isRoot:      true,
			}
			matchers = append(matchers, rootMatcher)
		}
	} else {
		// Pattern completed - add sub-matchers
		matchers = append(matchers, m.subMatchers...)

		// If recursive, restart the pattern
		if m.recursive {
			newMatcher := &RecursiveNodeMatcher{
				patterns:    m.patterns,
				currentStep: 0,
				handler:     m.handler,
				recursive:   m.recursive,
				subMatchers: m.subMatchers,
				isRoot:      false, // Recursive restarts are not root
			}
			matchers = append(matchers, newMatcher)
		}

		// If this is the root matcher, include itself for more sibling matching
		// but only if we're not recursive (to avoid double matching)
		if m.isRoot && !m.recursive {
			rootMatcher := &RecursiveNodeMatcher{
				patterns:    m.patterns,
				currentStep: 0,
				handler:     m.handler,
				recursive:   m.recursive,
				subMatchers: m.subMatchers,
				isRoot:      true,
			}
			matchers = append(matchers, rootMatcher)
		}
	}

	return matchers
}
