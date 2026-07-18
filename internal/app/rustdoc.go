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

		matched, parsed := parseDocsRsDocument(doc)
		// Nothing parsed: treat as not found instead of caching and returning
		// an empty document, which would mask docs.rs markup drift.
		if !matched || strings.TrimSpace(parsed) == "" {
			return "", 0, false, ErrNotFound
		}
		document = parsed

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

		matched, parsed := parseDocsRsDocument(doc)
		// Nothing parsed: treat as not found instead of caching and returning
		// an empty document, which would mask docs.rs markup drift.
		if !matched || strings.TrimSpace(parsed) == "" {
			return nil, ErrNotFound
		}
		document = parsed

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

// cleanRustDocSignature normalizes an impl/method signature extracted from a
// docs.rs code-header via RawInnerText. It drops the "ⓘ" notable-trait marker
// and flattens the internal newlines/tabs (e.g. from a where clause) into a
// single spaced line so the signature fits in an inline code span.
func cleanRustDocSignature(text string) string {
	text = strings.ReplaceAll(text, "ⓘ", "")
	text = strings.ReplaceAll(text, "\n", " ")
	text = strings.ReplaceAll(text, "\t", " ")
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

	// Struct fields (span.structfield) and their docs (a following div.docblock)
	// are direct children of section#main-content with no wrapping container,
	// unlike the overview docblock (nested in details.top-doc) and method
	// docblocks (nested in the impl lists). inFields, toggled by the section
	// headers, scopes the top-level div.docblock matcher to the Fields section so
	// it doesn't double-emit those other docblocks. fieldOpen pairs each field
	// signature with its optional following docblock.
	var inFields, fieldOpen bool

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

	// Declaration block: pre.item-decl holds the canonical signature/declaration
	// on item pages (struct, enum, fn, trait, type alias). Present on every item
	// page and absent on module index pages, so this is a no-op for the latter.
	itemDeclMatcher := dq.NewNodeMatcher(
		dq.NewMatchFunc("pre.item-decl"),
		func(n *html.Node) {
			decl := strings.TrimSpace(dq.RawInnerText(n, true))
			if decl != "" {
				builder.WriteString(fmt.Sprintf("```rust\n%s\n```\n", decl))
			}
		},
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
			// Track whether we're inside the Fields section so the top-level
			// div.docblock matcher only fires for field docs. Non-exhaustive
			// structs render the header as "Fields (Non-exhaustive)", so match on
			// the prefix rather than the exact string.
			inFields = strings.HasPrefix(text, "Fields")
			// Skip headers whose item lists we deliberately don't expand, so item
			// pages don't end with dangling empty sections.
			switch text {
			case "Auto Trait Implementations", "Blanket Implementations":
				return
			}
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

	// Impl blocks on item pages: each impl carries an h3.code-header ("impl Foo")
	// and its methods live in section.method > h4.code-header, each followed by a
	// div.docblock. Scope to the inherent-impl and trait-impl lists; the auto-trait
	// and blanket-impl lists are high-volume noise (Send/Sync/From<T>/…) and are
	// intentionally skipped.
	implHeaderMatcher := dq.NewNodeMatcher(
		dq.NewMatchFunc("h3.code-header"),
		func(n *html.Node) {
			sig := cleanRustDocSignature(dq.RawInnerText(n, true))
			if sig != "" {
				builder.WriteString(fmt.Sprintf("\n### %s\n", sig))
			}
		},
	)
	methodSigMatcher := dq.NewNodeMatcher(
		dq.NewMatchFunc("h4.code-header"),
		func(n *html.Node) {
			sig := cleanRustDocSignature(dq.RawInnerText(n, true))
			if sig != "" {
				builder.WriteString(fmt.Sprintf("- `%s`\n", sig))
			}
		},
	)
	methodDocMatcher := dq.NewNodeMatcher(
		dq.NewMatchFunc("div.docblock"),
		nil,
		dq.NewNodeMatcher(
			dq.NewMatchFunc("p"),
			func(n *html.Node) {
				text := strings.TrimSpace(dq.InnerText(n, true))
				if text != "" {
					builder.WriteString(fmt.Sprintf("  %s\n", text))
				}
			},
		),
	)
	implsMatcher := dq.NewNodeMatcher(
		dq.NewMatchFunc("div#implementations-list,div#trait-implementations-list"),
		nil,
		implHeaderMatcher,
		methodSigMatcher,
		methodDocMatcher,
	)

	// Enum variants: div.variants holds section.variant > h3.code-header (the
	// variant signature, e.g. "Number(Number)") each followed by a sibling
	// div.docblock. The "## Variants" header itself is already emitted by
	// sectionHeaderMatcher. Reuse methodDocMatcher for the variant docs since the
	// div.docblock > p shape is identical and the subtrees are disjoint.
	variantSigMatcher := dq.NewNodeMatcher(
		dq.NewMatchFunc("h3.code-header"),
		func(n *html.Node) {
			sig := cleanRustDocSignature(dq.RawInnerText(n, true))
			if sig != "" {
				builder.WriteString(fmt.Sprintf("- `%s`\n", sig))
			}
		},
	)
	variantsMatcher := dq.NewNodeMatcher(
		dq.NewMatchFunc("div.variants"),
		nil,
		variantSigMatcher,
		methodDocMatcher,
	)

	// Struct fields: span.structfield > code holds "name: Type"; a documented
	// field is followed by a sibling div.docblock. Both are direct children of
	// section#main-content, so fieldDocMatcher is gated on inFields (set by the
	// Fields section header) to avoid double-emitting the overview/method docblocks.
	structFieldMatcher := dq.NewNodeMatcher(
		dq.NewMatchFunc("span.structfield"),
		func(n *html.Node) {
			code := dq.FindOne(n, "code")
			if code == nil {
				return
			}
			sig := cleanRustDocSignature(dq.RawInnerText(code, true))
			if sig != "" {
				builder.WriteString(fmt.Sprintf("- `%s`\n", sig))
				fieldOpen = true
			}
		},
	)
	fieldDocMatcher := dq.NewNodeMatcher(
		dq.NewMatchFunc("div.docblock"),
		func(n *html.Node) {
			if !inFields || !fieldOpen {
				return
			}
			fieldOpen = false
			for _, p := range dq.FindAll(n, "p") {
				text := strings.TrimSpace(dq.InnerText(p, true))
				if text != "" {
					builder.WriteString(fmt.Sprintf("  %s\n", text))
				}
			}
		},
	)

	// Root: section#main-content
	var matched bool
	rootMatcher := dq.NewNodeMatcher(
		dq.NewMatchFunc("section#main-content"),
		func(n *html.Node) {
			matched = true
		},
		titleMatcher,
		itemDeclMatcher,
		overviewMatcher,
		sectionHeaderMatcher,
		itemTableMatcher,
		structFieldMatcher,
		fieldDocMatcher,
		variantsMatcher,
		implsMatcher,
	)

	dq.Traverse(doc, []dq.Matcher{rootMatcher})
	return matched, builder.String()
}
