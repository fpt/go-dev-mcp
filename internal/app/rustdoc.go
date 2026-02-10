package app

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/fpt/go-dev-mcp/internal/contentsearch"
	"github.com/fpt/go-dev-mcp/internal/infra"
	"github.com/fpt/go-dev-mcp/internal/model"
	"github.com/fpt/go-dev-mcp/pkg/dq"
	"github.com/patrickmn/go-cache"
	"github.com/pkg/errors"
	"golang.org/x/net/html"
)

func SearchRustDoc(httpcli *infra.HttpClient, query string) (string, error) {
	u := fmt.Sprintf(
		"https://docs.rs/releases/search?query=%s",
		url.QueryEscape(query),
	)
	bodyrdr, err := httpcli.HttpGet(u)
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

	matched, document := parseDocsRsSearchResult(doc)
	if !matched {
		return "", errors.New("no search results found")
	}

	return document, nil
}

// ReadRustDocPaged reads Rust documentation for a given crate URL with line-based paging.
// crateURL can be "serde", "serde/de", "tokio/runtime", etc.
func ReadRustDocPaged(
	httpcli *infra.HttpClient,
	crateURL string,
	offset, limit int,
) (string, int, bool, error) {
	cacheKey := fmt.Sprintf("rustdoc:%s", crateURL)

	var document string

	if cached, found := docCache.Get(cacheKey); found {
		document = cached.(string)
	} else {
		u := buildDocsRsURL(crateURL)
		bodyrdr, err := httpcli.HttpGet(u)
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

		_, document = parseDocsRsDocument(doc)

		docCache.Set(cacheKey, document, cache.DefaultExpiration)
	}

	lines := strings.Split(document, "\n")
	totalLines := len(lines)

	startIdx := offset
	if startIdx >= totalLines {
		return "", totalLines, false, nil
	}

	endIdx := startIdx + limit
	hasMore := endIdx < totalLines
	if endIdx > totalLines {
		endIdx = totalLines
	}

	pagedContent := strings.Join(lines[startIdx:endIdx], "\n")

	return pagedContent, totalLines, hasMore, nil
}

type RustDocSearchResult struct {
	CrateURL  string
	Matches   []model.SearchMatch
	Truncated bool
}

func SearchWithinRustDoc(
	httpcli *infra.HttpClient,
	crateURL string,
	keyword string,
	maxMatches int,
) (*RustDocSearchResult, error) {
	cacheKey := fmt.Sprintf("rustdoc:%s", crateURL)

	var document string

	if cached, found := docCache.Get(cacheKey); found {
		document = cached.(string)
	} else {
		u := buildDocsRsURL(crateURL)
		bodyrdr, err := httpcli.HttpGet(u)
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

		_, document = parseDocsRsDocument(doc)

		docCache.Set(cacheKey, document, cache.DefaultExpiration)
	}

	reader := strings.NewReader(document)
	matches, truncated, err := contentsearch.SearchInContent(reader, keyword, maxMatches)
	if err != nil {
		return nil, err
	}

	return &RustDocSearchResult{
		CrateURL:  crateURL,
		Matches:   matches,
		Truncated: truncated,
	}, nil
}

// buildDocsRsURL constructs the docs.rs URL from a crate URL.
// "serde" → "https://docs.rs/serde/latest/serde/"
// "serde/de" → "https://docs.rs/serde/latest/serde/de/"
// "serde-json" → "https://docs.rs/serde-json/latest/serde_json/"
func buildDocsRsURL(crateURL string) string {
	parts := strings.SplitN(crateURL, "/", 2)
	crateName := parts[0]
	// Rust crate names use hyphens but module paths use underscores
	crateModule := strings.ReplaceAll(crateName, "-", "_")

	if len(parts) == 1 {
		return fmt.Sprintf("https://docs.rs/%s/latest/%s/", crateName, crateModule)
	}

	return fmt.Sprintf(
		"https://docs.rs/%s/latest/%s/%s",
		crateName,
		crateModule,
		parts[1],
	)
}

// cleanRustDocHeading strips the § anchor symbol and "Copy item path" button text
// from docs.rs headings.
func cleanRustDocHeading(text string) string {
	text = strings.ReplaceAll(text, "§", "")
	text = strings.ReplaceAll(text, "Copy item path", "")
	// Collapse multiple spaces left by stripping
	for strings.Contains(text, "  ") {
		text = strings.ReplaceAll(text, "  ", " ")
	}
	return strings.TrimSpace(text)
}

func parseDocsRsSearchResult(doc *html.Node) (bool, string) {
	builder := strings.Builder{}

	var currentName string
	var currentHref string

	nameMatcher := dq.NewNodeMatcher(
		dq.NewMatchFunc("div.name"),
		func(n *html.Node) {
			currentName = strings.TrimSpace(dq.InnerText(n, true))
		},
	)
	descMatcher := dq.NewNodeMatcher(
		dq.NewMatchFunc("div.description"),
		func(n *html.Node) {
			desc := strings.TrimSpace(dq.InnerText(n, true))
			builder.WriteString(fmt.Sprintf("* %s\n", currentName))
			if currentHref != "" {
				builder.WriteString(fmt.Sprintf("\tURL: https://docs.rs%s\n", currentHref))
			}
			if desc != "" {
				builder.WriteString(fmt.Sprintf("\tDescription: %s\n", desc))
			}
		},
	)

	releaseMatcher := dq.NewNodeMatcher(
		dq.NewMatchFunc("a.release"),
		func(n *html.Node) {
			currentHref = dq.GetHref(n)
		},
		dq.NewNodeMatcher(
			dq.NewMatchFunc("div"),
			nil,
			nameMatcher,
			descMatcher,
		),
	)

	listMatcher := dq.NewNodeMatcher(
		dq.NewMatchFunc("ul"),
		nil,
		dq.NewNodeMatcher(
			dq.NewMatchFunc("li"),
			nil,
			releaseMatcher,
		),
	)

	var matched bool
	rootMatcher := dq.NewNodeMatcher(
		dq.NewMatchFunc("div.recent-releases-container"),
		func(n *html.Node) {
			matched = true
		},
		listMatcher,
	)

	dq.Traverse(doc, []dq.Matcher{rootMatcher})
	return matched, builder.String()
}

//nolint:funlen // Parser function with multiple matchers
func parseDocsRsDocument(doc *html.Node) (bool, string) {
	builder := strings.Builder{}

	// Title from div.main-heading > h1
	titleMatcher := dq.NewNodeMatcher(
		dq.NewMatchFunc("div.main-heading"),
		nil,
		dq.NewNodeMatcher(
			dq.NewMatchFunc("h1"),
			func(n *html.Node) {
				text := cleanRustDocHeading(dq.InnerText(n, true))
				builder.WriteString(fmt.Sprintf("# %s\n", text))
			},
		),
	)

	// Matchers for content inside div.docblock
	docblockHeaderMatcher := dq.NewNodeMatcher(
		dq.NewMatchFunc("h1,h2,h3,h4,h5,h6"),
		func(n *html.Node) {
			text := cleanRustDocHeading(dq.InnerText(n, true))
			h := strings.TrimPrefix(n.Data, "h")
			hn, _ := strconv.Atoi(h)
			builder.WriteString(
				fmt.Sprintf("\n%s %s\n", strings.Repeat("#", hn), text),
			)
		},
	)
	docblockPMatcher := dq.NewNodeMatcher(
		dq.NewMatchFunc("p"),
		func(n *html.Node) {
			text := strings.TrimSpace(dq.InnerText(n, true))
			if text != "" {
				builder.WriteString(fmt.Sprintf("%s\n", text))
			}
		},
	)
	docblockPreMatcher := dq.NewNodeMatcher(
		dq.NewMatchFunc("pre"),
		func(n *html.Node) {
			builder.WriteString(fmt.Sprintf("```\n%s\n```\n", dq.RawInnerText(n, true)))
		},
	)
	docblockListMatcher := dq.NewNodeMatcher(
		dq.NewMatchFunc("ul,ol"),
		nil,
		dq.NewNodeMatcher(
			dq.NewMatchFunc("li"),
			func(n *html.Node) {
				text := strings.TrimSpace(dq.InnerText(n, true))
				if text != "" {
					builder.WriteString(fmt.Sprintf("- %s\n", text))
				}
			},
		),
	)

	// Overview: details.top-doc > div.docblock
	overviewMatcher := dq.NewNodeMatcher(
		dq.NewMatchFunc("details.top-doc"),
		nil,
		dq.NewNodeMatcher(
			dq.NewMatchFunc("div.docblock"),
			nil,
			docblockHeaderMatcher,
			docblockPMatcher,
			docblockPreMatcher,
			docblockListMatcher,
		),
	)

	// Section headers: h2.section-header
	sectionHeaderMatcher := dq.NewNodeMatcher(
		dq.NewMatchFunc("h2.section-header"),
		func(n *html.Node) {
			text := cleanRustDocHeading(dq.InnerText(n, true))
			builder.WriteString(fmt.Sprintf("\n## %s\n", text))
		},
	)

	// Item tables: dl.item-table > dt/dd pairs
	var currentItem string
	dtMatcher := dq.NewNodeMatcher(
		dq.NewMatchFunc("dt"),
		func(n *html.Node) {
			currentItem = strings.TrimSpace(dq.InnerText(n, true))
		},
	)
	ddMatcher := dq.NewNodeMatcher(
		dq.NewMatchFunc("dd"),
		func(n *html.Node) {
			desc := strings.TrimSpace(dq.InnerText(n, true))
			if currentItem != "" {
				builder.WriteString(fmt.Sprintf("- %s: %s\n", currentItem, desc))
				currentItem = ""
			}
		},
	)
	itemTableMatcher := dq.NewNodeMatcher(
		dq.NewMatchFunc("dl.item-table"),
		nil,
		dtMatcher,
		ddMatcher,
	)

	// Root: section#main-content
	var matched bool
	rootMatcher := dq.NewNodeMatcher(
		dq.NewMatchFunc("section#main-content"),
		func(n *html.Node) {
			matched = true
		},
		titleMatcher,
		overviewMatcher,
		sectionHeaderMatcher,
		itemTableMatcher,
	)

	dq.Traverse(doc, []dq.Matcher{rootMatcher})
	return matched, builder.String()
}
