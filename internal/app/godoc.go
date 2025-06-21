package app

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/fpt/go-dev-mcp/internal/contentsearch"
	"github.com/fpt/go-dev-mcp/internal/infra"
	"github.com/fpt/go-dev-mcp/internal/model"
	"github.com/fpt/go-dev-mcp/pkg/dq"
	"github.com/patrickmn/go-cache"
	"github.com/pkg/errors"
	"golang.org/x/net/html"
)

const DefaultLinesPerPage = 100

var ErrNotFound = errors.New("not found")

// Global cache for storing parsed documentation
// TTL: 30 minutes, cleanup interval: 10 minutes
var docCache = cache.New(30*time.Minute, 10*time.Minute)

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

// ReadGoDocPaged reads Go documentation for a given package URL with line-based paging.
// packageURL must be in "golang.org/x/net/html" format.
// Returns: content, totalLines, hasMore, error
func ReadGoDocPaged(
	httpcli *infra.HttpClient,
	packageURL string,
	offset, limit int,
) (string, int, bool, error) {
	// Create cache key
	cacheKey := fmt.Sprintf("godoc:%s", packageURL)

	var document string

	// Check cache first
	if cached, found := docCache.Get(cacheKey); found {
		document = cached.(string)
	} else {
		// Cache miss - fetch and parse the document
		url := fmt.Sprintf("https://pkg.go.dev/%s", url.PathEscape(packageURL))
		bodyrdr, err := httpcli.HttpGet(url)
		if err != nil {
			return "", 0, false, errors.Wrap(err, "failed to make HTTP request")
		}
		if bodyrdr == nil {
			return "", 0, false, ErrNotFound
		}
		defer bodyrdr.Close()

		doc, err := html.Parse(bodyrdr)
		if err != nil {
			return "", 0, false, errors.Wrap(err, "failed to parse HTML")
		}

		// Get the full document first
		matched, parsedDoc := parseDocument(doc)

		// If no documentation found, try to parse the README section
		if !matched {
			_, parsedDoc = parseReadme(doc)
		}

		document = parsedDoc

		// Cache the parsed document for future requests
		docCache.Set(cacheKey, document, cache.DefaultExpiration)
	}

	// Split into lines and apply paging
	lines := strings.Split(document, "\n")
	totalLines := len(lines)

	// Apply offset and limit
	startIdx := offset
	if startIdx >= totalLines {
		return "", totalLines, false, nil
	}

	endIdx := startIdx + limit
	hasMore := endIdx < totalLines
	if endIdx > totalLines {
		endIdx = totalLines
	}

	// Join the selected lines back together
	pagedContent := strings.Join(lines[startIdx:endIdx], "\n")

	return pagedContent, totalLines, hasMore, nil
}

type GoDocSearchResult struct {
	PackageURL string
	Matches    []model.SearchMatch
	Truncated  bool
}

// SearchWithinGoDoc searches for a keyword within Go documentation and returns all matches.
// Similar to SearchLocalFiles but for a single Go documentation page.
func SearchWithinGoDoc(
	httpcli *infra.HttpClient,
	packageURL string,
	keyword string,
	maxMatches int,
) (*GoDocSearchResult, error) {
	// Create cache key
	cacheKey := fmt.Sprintf("godoc:%s", packageURL)

	var document string

	// Check cache first
	if cached, found := docCache.Get(cacheKey); found {
		document = cached.(string)
	} else {
		// Cache miss - fetch and parse the document
		url := fmt.Sprintf("https://pkg.go.dev/%s", url.PathEscape(packageURL))
		bodyrdr, err := httpcli.HttpGet(url)
		if err != nil {
			return nil, errors.Wrap(err, "failed to make HTTP request")
		}
		if bodyrdr == nil {
			return nil, ErrNotFound
		}
		defer bodyrdr.Close()

		doc, err := html.Parse(bodyrdr)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse HTML")
		}

		// Get the full document first
		matched, parsedDoc := parseDocument(doc)

		// If no documentation found, try to parse the README section
		if !matched {
			_, parsedDoc = parseReadme(doc)
		}

		document = parsedDoc

		// Cache the parsed document for future requests
		docCache.Set(cacheKey, document, cache.DefaultExpiration)
	}

	// Search through the document using the shared contentsearch package
	reader := strings.NewReader(document)
	matches, truncated, err := contentsearch.SearchInContent(reader, keyword, maxMatches)
	if err != nil {
		return nil, err
	}

	return &GoDocSearchResult{
		PackageURL: packageURL,
		Matches:    matches,
		Truncated:  truncated,
	}, nil
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
			builder.WriteString(
				fmt.Sprintf("\n%s %s\n", strings.Repeat("#", hn), dq.InnerText(n, true)),
			)
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
			// section is a container of multiple divs.
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
			builder.WriteString(
				fmt.Sprintf("\n%s %s\n", strings.Repeat("#", hn), dq.InnerText(n, true)),
			)
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
