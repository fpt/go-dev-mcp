package app

import (
	"fmt"
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

type pyModEntry struct {
	Name        string
	Description string
	URL         string
}

func SearchPyDoc(httpcli *infra.HttpClient, query string) (string, error) {
	cacheKey := "pydoc:modindex"

	var entries []pyModEntry

	if cached, found := docCache.Get(cacheKey); found {
		entries = cached.([]pyModEntry)
	} else {
		u := "https://docs.python.org/3/py-modindex.html"
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

		matched, parsed := parsePyModIndex(doc)
		if !matched {
			return "", errors.New("no module index found")
		}

		entries = parsed
		docCache.Set(cacheKey, entries, cache.DefaultExpiration)
	}

	// Filter entries by query (case-insensitive)
	queryLower := strings.ToLower(query)
	builder := strings.Builder{}
	count := 0

	for _, entry := range entries {
		if strings.Contains(strings.ToLower(entry.Name), queryLower) ||
			strings.Contains(strings.ToLower(entry.Description), queryLower) {
			builder.WriteString(fmt.Sprintf("* %s\n", entry.Name))
			builder.WriteString(fmt.Sprintf("\tURL: %s\n", entry.URL))
			if entry.Description != "" {
				builder.WriteString(fmt.Sprintf("\tDescription: %s\n", entry.Description))
			}
			count++
		}
	}

	if count == 0 {
		return "", errors.New("no matching modules found")
	}

	return fmt.Sprintf("Found %d matching Python modules:\n%s", count, builder.String()), nil
}

func ReadPyDocPaged(
	httpcli *infra.HttpClient,
	moduleName string,
	offset, limit int,
) (string, int, bool, error) {
	cacheKey := fmt.Sprintf("pydoc:%s", moduleName)

	var document string

	if cached, found := docCache.Get(cacheKey); found {
		document = cached.(string)
	} else {
		u := buildPyDocURL(moduleName)
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

		_, document = parsePyDocPage(doc)

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

type PyDocSearchResult struct {
	ModuleName string
	Matches    []model.SearchMatch
	Truncated  bool
}

func SearchWithinPyDoc(
	httpcli *infra.HttpClient,
	moduleName string,
	keyword string,
	maxMatches int,
) (*PyDocSearchResult, error) {
	cacheKey := fmt.Sprintf("pydoc:%s", moduleName)

	var document string

	if cached, found := docCache.Get(cacheKey); found {
		document = cached.(string)
	} else {
		u := buildPyDocURL(moduleName)
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

		_, document = parsePyDocPage(doc)

		docCache.Set(cacheKey, document, cache.DefaultExpiration)
	}

	reader := strings.NewReader(document)
	matches, truncated, err := contentsearch.SearchInContent(reader, keyword, maxMatches)
	if err != nil {
		return nil, err
	}

	return &PyDocSearchResult{
		ModuleName: moduleName,
		Matches:    matches,
		Truncated:  truncated,
	}, nil
}

// buildPyDocURL constructs the docs.python.org URL from a module name.
// "json" → "https://docs.python.org/3/library/json.html"
// "os.path" → "https://docs.python.org/3/library/os.path.html"
func buildPyDocURL(moduleName string) string {
	return fmt.Sprintf("https://docs.python.org/3/library/%s.html", moduleName)
}

// cleanPyDocHeading strips the ¶ anchor symbol from Python docs headings.
func cleanPyDocHeading(text string) string {
	text = strings.ReplaceAll(text, "¶", "")
	for strings.Contains(text, "  ") {
		text = strings.ReplaceAll(text, "  ", " ")
	}
	return strings.TrimSpace(text)
}

// cleanPyDocSignature extracts clean signature text from a dt element's raw text.
func cleanPyDocSignature(text string) string {
	text = strings.ReplaceAll(text, "¶", "")
	text = strings.ReplaceAll(text, "[source]", "")
	for strings.Contains(text, "  ") {
		text = strings.ReplaceAll(text, "  ", " ")
	}
	return strings.TrimSpace(text)
}

func parsePyModIndex(doc *html.Node) (bool, []pyModEntry) {
	var entries []pyModEntry

	trMatcher := dq.NewNodeMatcher(
		dq.NewMatchFunc("tr"),
		func(n *html.Node) {
			entry := extractModuleFromRow(n)
			if entry != nil {
				entries = append(entries, *entry)
			}
		},
	)

	var matched bool
	rootMatcher := dq.NewNodeMatcher(
		dq.NewMatchFunc("table.modindextable"),
		func(n *html.Node) {
			matched = true
		},
		trMatcher,
	)

	dq.Traverse(doc, []dq.Matcher{rootMatcher})
	return matched, entries
}

// extractModuleFromRow extracts a module entry from a table row in the module index.
// Returns nil for non-module rows (caption/padding rows).
func extractModuleFromRow(tr *html.Node) *pyModEntry {
	tds := dq.FindAll(tr, "td")
	if len(tds) < 3 {
		return nil
	}

	// Second td (index 1): module link
	a := dq.FindDirectChild(tds[1], "a[href*=#module-]")
	if a == nil {
		return nil // Not a module row (cap/pcap row)
	}
	moduleName := strings.TrimSpace(dq.InnerText(a, true))
	if moduleName == "" {
		return nil
	}
	moduleHref := dq.GetHref(a)

	// Third td (index 2): description
	description := strings.TrimSpace(dq.InnerText(tds[2], true))

	return &pyModEntry{
		Name:        moduleName,
		Description: description,
		URL:         fmt.Sprintf("https://docs.python.org/3/%s", moduleHref),
	}
}

//nolint:funlen // Parser function with multiple matchers
func parsePyDocPage(doc *html.Node) (bool, string) {
	builder := strings.Builder{}

	// h1 matcher for page title
	h1Matcher := dq.NewNodeMatcher(
		dq.NewMatchFunc("h1"),
		func(n *html.Node) {
			text := cleanPyDocHeading(dq.InnerText(n, true))
			builder.WriteString(fmt.Sprintf("# %s\n\n", text))
		},
	)

	// Heading matcher for h2, h3, h4
	headingMatcher := dq.NewNodeMatcher(
		dq.NewMatchFunc("h2,h3,h4"),
		func(n *html.Node) {
			text := cleanPyDocHeading(dq.InnerText(n, true))
			h := strings.TrimPrefix(n.Data, "h")
			hn, _ := strconv.Atoi(h)
			builder.WriteString(
				fmt.Sprintf("\n%s %s\n", strings.Repeat("#", hn), text),
			)
		},
	)

	// Paragraph matcher
	pMatcher := dq.NewNodeMatcher(
		dq.NewMatchFunc("p"),
		func(n *html.Node) {
			text := strings.TrimSpace(dq.InnerText(n, true))
			if text != "" {
				builder.WriteString(fmt.Sprintf("%s\n", text))
			}
		},
	)

	// Code block matcher
	preMatcher := dq.NewNodeMatcher(
		dq.NewMatchFunc("pre"),
		func(n *html.Node) {
			code := strings.TrimRight(dq.RawInnerText(n, true), "\n")
			if code != "" {
				builder.WriteString(fmt.Sprintf("```\n%s\n```\n", code))
			}
		},
	)

	// Signature matcher (dt.sig-object inside dl.py)
	dtMatcher := dq.NewNodeMatcher(
		dq.NewMatchFunc("dt.sig-object"),
		func(n *html.Node) {
			sig := cleanPyDocSignature(dq.RawInnerText(n, true))
			builder.WriteString(fmt.Sprintf("\n%s\n", sig))
		},
	)

	// Inner dd matcher (for methods inside class dd - one level of nesting)
	innerDdMatcher := dq.NewNodeMatcher(
		dq.NewMatchFunc("dd"),
		nil,
		pMatcher,
		preMatcher,
	)

	// Inner dl matcher (for methods inside class dd)
	innerDlMatcher := dq.NewNodeMatcher(
		dq.NewMatchFunc("dl.py"),
		nil,
		dtMatcher,
		innerDdMatcher,
	)

	// Outer dd matcher
	ddMatcher := dq.NewNodeMatcher(
		dq.NewMatchFunc("dd"),
		nil,
		pMatcher,
		preMatcher,
		innerDlMatcher,
	)

	// Outer dl matcher (top-level definitions)
	dlMatcher := dq.NewNodeMatcher(
		dq.NewMatchFunc("dl.py"),
		nil,
		dtMatcher,
		ddMatcher,
	)

	// Section matcher for nested sections
	sectionMatcher := dq.NewNodeMatcher(
		dq.NewMatchFunc("section"),
		nil,
		headingMatcher,
	)

	// Root matcher
	var matched bool
	rootMatcher := dq.NewNodeMatcher(
		dq.NewMatchFunc("section#module-*"),
		func(n *html.Node) {
			matched = true
		},
		h1Matcher,
		headingMatcher,
		sectionMatcher,
		dlMatcher,
		pMatcher,
		preMatcher,
	)

	dq.Traverse(doc, []dq.Matcher{rootMatcher})
	return matched, builder.String()
}
