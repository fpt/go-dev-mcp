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
// [=]: match attribute value e.g. "div[role=note]"
// [*=]: match attribute contains e.g. "a[href*=#module-]"
// [^=]: match attribute prefix e.g. "a[href^=/3/library/]"
// [$=]: match attribute suffix e.g. "a[href$=.html]"
// #: match id e.g. "div#id"
// .: match class e.g. "div.class"
// >: match direct child e.g. ">span" (RecursiveNodeMatcher only)
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
			ss := strings.SplitN(pattern, "[", 2)
			element, attrExpr := ss[0], ss[1]
			attrExpr = attrExpr[:len(attrExpr)-1] // strip trailing ']'
			if (element == "" || n.Data == element) && matchAttrExpr(n, attrExpr) {
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

// matchAttrExpr handles attribute selectors including value matching.
// Supported forms:
//
//	[attr]        — attribute exists
//	[attr=val]    — exact value match
//	[attr*=val]   — value contains substring
//	[attr^=val]   — value starts with prefix
//	[attr$=val]   — value ends with suffix
func matchAttrExpr(n *html.Node, expr string) bool {
	// Try value operators in order: *=, ^=, $=, =
	for _, op := range []string{"*=", "^=", "$=", "="} {
		idx := strings.Index(expr, op)
		if idx < 0 {
			continue
		}
		key := expr[:idx]
		val := expr[idx+len(op):]
		return matchAttrValue(n, key, op, val)
	}
	// No operator — existence check
	return hasAttr(n, expr)
}

func matchAttrValue(n *html.Node, key, op, val string) bool {
	for _, attr := range n.Attr {
		if attr.Key != key {
			continue
		}
		switch op {
		case "=":
			return attr.Val == val
		case "*=":
			return strings.Contains(attr.Val, val)
		case "^=":
			return strings.HasPrefix(attr.Val, val)
		case "$=":
			return strings.HasSuffix(attr.Val, val)
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

// FindOne returns the first element in the subtree matching the selector,
// or nil if none is found. Searches the entire subtree depth-first.
func FindOne(root *html.Node, selector string) *html.Node {
	matchFn := NewMatchFunc(selector)
	return findOneRecursive(root, matchFn)
}

func findOneRecursive(n *html.Node, matchFn MatchFunc) *html.Node {
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode && matchFn(c) {
			return c
		}
		if found := findOneRecursive(c, matchFn); found != nil {
			return found
		}
	}
	return nil
}

// FindAll returns all elements in the subtree matching the selector.
// Searches the entire subtree depth-first.
func FindAll(root *html.Node, selector string) []*html.Node {
	matchFn := NewMatchFunc(selector)
	var results []*html.Node
	findAllRecursive(root, matchFn, &results)
	return results
}

func findAllRecursive(n *html.Node, matchFn MatchFunc, results *[]*html.Node) {
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode && matchFn(c) {
			*results = append(*results, c)
		}
		findAllRecursive(c, matchFn, results)
	}
}

// FindDirectChild returns the first direct child element matching the selector,
// or nil if none is found. Only checks immediate children, not deeper descendants.
func FindDirectChild(n *html.Node, selector string) *html.Node {
	matchFn := NewMatchFunc(selector)
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode && matchFn(c) {
			return c
		}
	}
	return nil
}

// GetAttr returns the value of the named attribute, or "" if not found.
func GetAttr(n *html.Node, key string) string {
	for _, attr := range n.Attr {
		if attr.Key == key {
			return strings.TrimSpace(attr.Val)
		}
	}
	return ""
}

// GetHref is a convenience shorthand for GetAttr(n, "href").
func GetHref(n *html.Node) string {
	return GetAttr(n, "href")
}

// NodeFilter returns true to include a child node, false to skip it.
type NodeFilter func(n *html.Node) bool

// DefaultNodeFilter skips anchor elements with aria-label attributes
// (headerlink anchors commonly found on documentation sites).
func DefaultNodeFilter(n *html.Node) bool {
	if n.Type == html.ElementNode && n.Data == "a" && hasAttr(n, "aria-label") {
		return false
	}
	return true
}

// InnerText extracts text content from a node using the default filter.
func InnerText(n *html.Node, recurse bool) string {
	return InnerTextWithFilter(n, recurse, DefaultNodeFilter)
}

// InnerTextWithFilter extracts text content from a node, skipping child nodes
// for which filter returns false. When recurse is true, element children are
// recursively processed and surrounded by spaces.
func InnerTextWithFilter(n *html.Node, recurse bool, filter NodeFilter) string {
	var text string
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode && !filter(c) {
			continue
		}

		if c.Type == html.TextNode {
			text += strings.TrimSpace(c.Data)
		}

		if recurse && c.Type == html.ElementNode {
			text += fmt.Sprintf(" %s ", InnerTextWithFilter(c, recurse, filter))
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
