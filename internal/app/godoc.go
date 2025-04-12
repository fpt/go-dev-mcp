package app

import (
	"fmt"
	"strings"

	"fujlog.net/godev-mcp/internal/infra"
	"fujlog.net/godev-mcp/pkg/htmlu"
	"github.com/pkg/errors"
	"golang.org/x/net/html"
)

func SearchGoDoc(httpcli *infra.HttpClient, query string) (string, error) {
	url := fmt.Sprintf("https://pkg.go.dev/search?q=%s", query)
	bodyrdr, err := httpcli.HttpGet(url)
	if err != nil {
		return "", errors.Wrap(err, "failed to make HTTP request")
	}
	defer bodyrdr.Close()

	doc, err := html.Parse(bodyrdr)
	if err != nil {
		return "", errors.Wrap(err, "failed to parse HTML")
	}

	// find all <div class="SearchSnippet"> elements
	var results []SingleResult

	// Search for search results in the document
	htmlu.Walk(doc, func(n *html.Node) bool {
		if n.Type == html.ElementNode && n.Data == "div" && htmlu.HasClass(n, "SearchSnippet") {
			r := processSingleResult(n)
			results = append(results, r)
			return false // Stop walking after finding the first matching element
		}
		return true
	})

	var resultStrings []string
	for _, result := range results {
		resultStrings = append(resultStrings, fmt.Sprintf("* %s\n\tURL: %s\n\tDescription: %s", result.Name, result.URL, result.Description))
	}

	return strings.Join(resultStrings, "\n"), nil
}

type SingleResult struct {
	Name        string
	URL         string
	Description string
}

func processSingleResult(p *html.Node) SingleResult {
	result := SingleResult{}

	// Process direct header container
	htmlu.Walk(p, func(n *html.Node) bool {
		if n.Type == html.ElementNode && n.Data == "div" && htmlu.HasClass(n, "SearchSnippet-headerContainer") {
			// Look for anchor inside header container
			for a := range n.Descendants() {
				if a.Type == html.ElementNode && a.Data == "a" {
					result.Name = htmlu.GetText(a, false)
					url := htmlu.GetHref(a)
					if url != "" {
						// Remove the leading slash if it exists
						if strings.HasPrefix(url, "/") {
							url = url[1:]
						}
						result.URL = url
					}
					break // Found what we need
				}
			}
			return false // Stop walking after finding the first matching element
		}

		// Look for description in synopsis paragraph
		if n.Type == html.ElementNode && n.Data == "p" && htmlu.HasClass(n, "SearchSnippet-synopsis") {
			result.Description = htmlu.GetText(n, true)
		}
		return true // Continue walking
	})

	return result
}

// ReadGoDoc reads Go documentation for a given package URL.
// packageURL must be in "golang.org/x/net/html" format.
func ReadGoDoc(httpcli *infra.HttpClient, packageURL string) (string, error) {
	url := fmt.Sprintf("https://pkg.go.dev/%s", packageURL)
	bodyrdr, err := httpcli.HttpGet(url)
	if err != nil {
		return "", errors.Wrap(err, "failed to make HTTP request")
	}
	defer bodyrdr.Close()

	doc, err := html.Parse(bodyrdr)
	if err != nil {
		return "", errors.Wrap(err, "failed to parse HTML")
	}

	var documents []string
	for n := range doc.Descendants() {
		if n.Type == html.ElementNode && n.Data == "div" && htmlu.HasClass(n, "UnitDoc") {
			htmlu.Walk(n, func(n *html.Node) bool {
				if n.Type == html.ElementNode && n.Data == "section" && htmlu.HasClass(n, "Documentation-*") {
					documents = append(documents, processSection(n))
					return false // Stop walking after finding the first matching section
				}
				return true
			})
			break
		}
	}

	document := ""
	for _, doc := range documents {
		document += fmt.Sprintf("%s\n", doc)
	}

	return document, nil
}

func processSection(p *html.Node) string {
	var documentation string
	for c := p.FirstChild; c != nil; c = c.NextSibling {
		documentation += processSubsection(c)
	}
	return documentation
}

func processSubsection(n *html.Node) string {
	var documentation string
	if n.Type == html.ElementNode {
		switch n.Data {
		case "h1", "h2", "h3", "h4", "h5", "h6":
			headers := []string{"# ", "## ", "### ", "#### ", "##### ", "###### "}
			level := strings.Index("h1h2h3h4h5h6", n.Data)
			if level != -1 {
				documentation += fmt.Sprintf("%s%s\n", headers[level/2], strings.TrimSpace(htmlu.GetRawText(n, true)))
			} else {
				documentation += fmt.Sprintf("%s\n", strings.TrimSpace(htmlu.GetRawText(n, true)))
			}
		case "div":
			if htmlu.HasClass(n, "Documentation-*") {
				documentation += processSection(n)
			}
		case "p":
			documentation += fmt.Sprintf("%s\n", htmlu.GetText(n, true))
		case "pre":
			documentation += fmt.Sprintf("```\n%s\n```\n", htmlu.GetText(n, true))
		case "code":
			documentation += fmt.Sprintf("`%s`\n", htmlu.GetText(n, true))
		case "details":
			documentation += processDetails(n)
		case "a":
			href := htmlu.GetHref(n)
			if href != "" {
				documentation += fmt.Sprintf("[%s](%s)\n", htmlu.GetText(n, true), href)
			} else {
				documentation += fmt.Sprintf("%s\n", htmlu.GetText(n, true))
			}
		case "ul", "ol":
			documentation += processList(n, 0)
		case "dl":
			for c := range n.Descendants() {
				if c.Type == html.ElementNode && c.Data == "dt" {
					documentation += fmt.Sprintf("- %s\n", htmlu.GetText(c, true))
				} else if c.Type == html.ElementNode && c.Data == "dd" {
					documentation += fmt.Sprintf("  %s\n", htmlu.GetText(c, true))
				}
			}
		}
	}
	return documentation
}

func processList(n *html.Node, level int) string {
	var documentation string
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode && (c.Data == "li" || c.Data == "dt") {
			listNodes := htmlu.SelectChildNodes(c, "ul", "ol")
			fmt.Printf("listNodes: %+v\n", listNodes)
			if len(listNodes) > 0 {
				for _, list := range listNodes {
					documentation += processList(list, level+1)
				}
			} else {
				documentation += fmt.Sprintf("%s- %s\n", strings.Repeat(" ", level*2), htmlu.GetText(c, true))
			}
		}
	}
	return documentation
}

func processDetails(n *html.Node) string {
	var documentation string
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode && c.Data == "summary" {
			documentation += fmt.Sprintf("** %s **\n", htmlu.GetRawText(c, true))
		} else if c.Type == html.ElementNode && htmlu.HasClass(c, "Documentation-exampleDetailsBody") {
			for cc := c.FirstChild; cc != nil; cc = cc.NextSibling {
				if cc.Type == html.ElementNode && (cc.Data == "textarea") {
					documentation += fmt.Sprintf("```\n%s\n```\n", htmlu.GetRawText(cc, true))
				} else if cc.Type == html.ElementNode && cc.Data == "pre" {
					documentation += fmt.Sprintf("```\n%s\n```\n", htmlu.GetRawText(cc, true))
				}
			}
		}
	}
	return documentation
}
