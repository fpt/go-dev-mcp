package htmlu

import (
	"path/filepath"
	"strings"

	"golang.org/x/net/html"
)

type NodeWalkFunc func(*html.Node) bool

func Walk(p *html.Node, f NodeWalkFunc) {
	for c := p.FirstChild; c != nil; c = c.NextSibling {
		into := f(c)
		if c.Type == html.ElementNode && into {
			Walk(c, f)
		}
	}
}

func HasClass(n *html.Node, pattern string) bool {
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

func HasAttr(n *html.Node, pattern string) bool {
	for _, attr := range n.Attr {
		match, err := filepath.Match(pattern, attr.Key)
		if err == nil && match {
			return true
		}
	}
	return false
}

func SelectChildNodes(n *html.Node, patterns ...string) []*html.Node {
	var nodes []*html.Node
	Walk(n, func(n *html.Node) bool {
		for _, pattern := range patterns {
			if n.Type == html.ElementNode && n.Data == pattern {
				nodes = append(nodes, n)
			}
		}
		return true
	})
	return nodes
}

func GetHref(n *html.Node) string {
	for _, attr := range n.Attr {
		if attr.Key == "href" {
			return strings.TrimSpace(attr.Val)
		}
	}
	return ""
}

func GetText(n *html.Node, recurse bool) string {
	if n.Type == html.TextNode {
		return strings.TrimSpace(n.Data)
	}

	var text string
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode && c.Data == "a" && HasAttr(c, "aria-label") {
			continue
		}

		text += GetText(c, recurse)
	}
	return text
}

func GetRawText(n *html.Node, recurse bool) string {
	if n.Type == html.TextNode {
		return n.Data
	}

	var text string
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode && c.Data == "a" && HasAttr(c, "aria-label") {
			continue
		}

		text += GetRawText(c, recurse)
	}
	return text
}
