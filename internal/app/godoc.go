package app

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"fujlog.net/godev-mcp/internal/infra"
	"fujlog.net/godev-mcp/pkg/dq"
	"github.com/pkg/errors"
	"golang.org/x/net/html"
)

var ErrNotFound = errors.New("not found")

func SearchGoDoc(httpcli *infra.HttpClient, query string) (string, error) {
	url := fmt.Sprintf("https://pkg.go.dev/search?q=%s", url.QueryEscape(query))
	bodyrdr, err := httpcli.HttpGet(url)
	if err != nil {
		return "", errors.Wrap(err, "failed to make HTTP request")
	}
	if bodyrdr == nil {
		return "", ErrNotFound
	}
	defer bodyrdr.Close()

	doc, err := html.Parse(bodyrdr)
	if err != nil {
		return "", errors.Wrap(err, "failed to parse HTML")
	}
	matched, document := parseSearchResult(doc)
	if !matched {
		matched, readme := parseReadme(doc)
		if matched {
			return readme, nil
		}

		_, document = parseDocument(doc)
	}

	return document, nil
}

// ReadGoDoc reads Go documentation for a given package URL.
// packageURL must be in "golang.org/x/net/html" format.
func ReadGoDoc(httpcli *infra.HttpClient, packageURL string) (string, error) {
	url := fmt.Sprintf("https://pkg.go.dev/%s", url.PathEscape(packageURL))
	bodyrdr, err := httpcli.HttpGet(url)
	if err != nil {
		return "", errors.Wrap(err, "failed to make HTTP request")
	}
	if bodyrdr == nil {
		return "", ErrNotFound
	}
	defer bodyrdr.Close()

	doc, err := html.Parse(bodyrdr)
	if err != nil {
		return "", errors.Wrap(err, "failed to parse HTML")
	}

	matched, document := parseDocument(doc)

	// If no documentation found, try to parse the README section
	if !matched {
		_, readme := parseReadme(doc)
		return readme, nil
	}

	return document, nil
}

func parseSearchResult(doc *html.Node) (bool, string) {
	builder := strings.Builder{}

	headerMatcher := dq.NewNodeMatcher(
		dq.NewMatchFunc("h2"),
		func(n *html.Node) {
			builder.WriteString(fmt.Sprintf("* %s\n", dq.InnerText(n, true)))
		},
		dq.NewNodeMatcher(
			dq.NewMatchFunc("span"),
			func(n *html.Node) {
				url := dq.InnerText(n, false)
				url = strings.Trim(url, "()")
				builder.WriteString(fmt.Sprintf("\tURL: %s\n", url))
			},
		),
	)
	pMatcher := dq.NewNodeMatcher(
		dq.NewMatchFunc("p"),
		func(n *html.Node) {
			builder.WriteString(fmt.Sprintf("\tDescription: %s\n", dq.InnerText(n, true)))
		},
	)
	snippetMatcher := dq.NewNodeMatcher(
		dq.NewMatchFunc("div.SearchSnippet"),
		nil,
		headerMatcher,
		pMatcher,
	)

	var matched bool
	rootMatcher := dq.NewNodeMatcher(
		dq.NewMatchFunc("div.SearchResults"),
		func(n *html.Node) {
			matched = true
		},
		dq.NewNodeMatcher(
			dq.NewMatchFunc("div.SearchResults-summary"),
			func(n *html.Node) {
				builder.WriteString(fmt.Sprintf("Summary: %s\n", dq.InnerText(n, true)))
			},
		),
		dq.NewNodeMatcher(
			dq.NewMatchFunc("div"),
			nil,
			snippetMatcher,
		),
	)

	dq.Traverse(doc, []dq.Matcher{rootMatcher})
	return matched, builder.String()
}

func parseDocument(doc *html.Node) (bool, string) {
	builder := strings.Builder{}

	headerMatcher := dq.NewNodeMatcher(
		dq.NewMatchFunc("h1,h2,h3,h4,h5,h6"),
		func(n *html.Node) {
			h := strings.TrimPrefix(n.Data, "h")
			hn, _ := strconv.Atoi(h)
			builder.WriteString(fmt.Sprintf("\n%s %s\n", strings.Repeat("#", hn), dq.InnerText(n, true)))
		},
	)
	listMatcher := dq.NewNodeMatcher(
		dq.NewMatchFunc("ul,ol"),
		nil,
		dq.NewNodeMatcher(
			dq.NewMatchFunc("li"),
			func(n *html.Node) {
				if dq.HasChild(n, "a") {
					builder.WriteString(fmt.Sprintf("- %s\n", dq.InnerText(n, true)))
				}
			},
			dq.NewNodeMatcher(
				dq.NewMatchFunc("ul,ol"),
				nil,
				dq.NewNodeMatcher(
					dq.NewMatchFunc("li"),
					func(n *html.Node) {
						builder.WriteString(fmt.Sprintf("    - %s\n", dq.InnerText(n, true)))
					},
				),
			),
		),
	)
	pMatcher := dq.NewNodeMatcher(
		dq.NewMatchFunc("p"),
		func(n *html.Node) {
			builder.WriteString(fmt.Sprintf("%s\n", dq.InnerText(n, true)))
		},
	)
	preMatcher := dq.NewNodeMatcher(
		dq.NewMatchFunc("pre"),
		func(n *html.Node) {
			builder.WriteString(fmt.Sprintf("```\n%s\n```\n", dq.RawInnerText(n, true)))
		},
	)
	sectionMatcher := dq.NewNodeMatcher(
		dq.NewMatchFunc("section"),
		func(n *html.Node) {
			builder.WriteString("Found section\n")
		},
		headerMatcher,
		dq.NewNodeMatcher(
			dq.NewMatchFunc("div"),
			nil,
			listMatcher,
			headerMatcher,
			pMatcher,
			preMatcher,
		),
		pMatcher,
		preMatcher,
	)

	var matched bool
	rootMatcher := dq.NewNodeMatcher(
		dq.NewMatchFunc("div.Documentation"),
		func(n *html.Node) {
			matched = true
		},
		dq.NewNodeMatcher(
			dq.NewMatchFunc("div"),
			nil,
			headerMatcher,
			sectionMatcher,
			listMatcher,
		),
	)

	dq.Traverse(doc, []dq.Matcher{rootMatcher})
	return matched, builder.String()
}

func parseReadme(doc *html.Node) (bool, string) {
	builder := strings.Builder{}

	headerMatcher := dq.NewNodeMatcher(
		dq.NewMatchFunc("h1,h2,h3,h4,h5,h6"),
		func(n *html.Node) {
			h := strings.TrimPrefix(n.Data, "h")
			hn, _ := strconv.Atoi(h)
			builder.WriteString(fmt.Sprintf("\n%s %s\n", strings.Repeat("#", hn), dq.InnerText(n, true)))
		},
	)
	listMatcher := dq.NewNodeMatcher(
		dq.NewMatchFunc("ul,ol"),
		nil,
		dq.NewNodeMatcher(
			dq.NewMatchFunc("li"),
			func(n *html.Node) {
				builder.WriteString(fmt.Sprintf("- %s\n", dq.InnerText(n, true)))
			},
		),
	)
	pMatcher := dq.NewNodeMatcher(
		dq.NewMatchFunc("p"),
		func(n *html.Node) {
			builder.WriteString(fmt.Sprintf("%s\n", dq.InnerText(n, true)))
		},
	)
	preMatcher := dq.NewNodeMatcher(
		dq.NewMatchFunc("pre"),
		func(n *html.Node) {
			builder.WriteString(fmt.Sprintf("```\n%s\n```\n", dq.RawInnerText(n, true)))
		},
	)

	var matched bool
	rootMatcher := dq.NewNodeMatcher(
		dq.NewMatchFunc("div.Overview-readmeContent"),
		func(n *html.Node) {
			matched = true
		},
		headerMatcher,
		listMatcher,
		pMatcher,
		preMatcher,
	)

	dq.Traverse(doc, []dq.Matcher{rootMatcher})
	return matched, builder.String()
}
