---
name: add-doc-integration
description: |
  This skill should be used when the user asks to add support for a new
  documentation site or integrate a new doc source. Trigger phrases include
  "add Java docs support", "integrate Node.js docs", "add MDN support",
  "new documentation site integration", "add docs for Ruby", "support reading
  docs from [site]". It guides implementing the 3-tool pattern
  (search/read/search_within) using the dq HTML document query package,
  following the proven pattern from the existing godoc (pkg.go.dev),
  rustdoc (docs.rs), and pydoc (docs.python.org) integrations.
---

# Add a New Documentation Site Integration

You are adding a new documentation site to the go-dev-mcp MCP server. This codebase already has three working integrations (godoc, rustdoc, pydoc) that all follow the same pattern. Your job is to replicate this pattern for the new site.

## Overview

Each documentation site integration consists of exactly **3 MCP tools**:
1. `search_{name}` - Search/discover documents (modules, packages, crates, etc.)
2. `read_{name}` - Read a specific document page with pagination
3. `search_within_{name}` - Keyword search within a specific document page

These tools are implemented across **4 layers** + golden tests.

## Step-by-step Implementation

### Phase 1: Explore the target documentation site

Before writing any code, you MUST understand the HTML structure of the target site.

1. **Identify 2 key page types**:
   - **Search/index page**: Where users discover available documents (e.g., module index, package listing)
   - **Document page**: Individual documentation page (e.g., a module reference, API docs)

2. **Fetch sample HTML** and inspect the DOM structure. Look for:
   - Root container element (e.g., `section#module-*`, `div.rustdoc`, `article`)
   - Headings (h1-h4) that structure the page
   - Definition lists or similar patterns for API entries (functions, classes, methods)
   - Code blocks (pre, code elements)
   - Signature elements (function/method signatures)
   - Navigation/link structure in the search/index page

3. **Verify dq compatibility**. The `pkg/dq` package supports:
   - Tag matching: `"div"`, `"section"`, `"dl"`
   - Class matching: `"dl.py"` matches `<dl class="py function">` (checks individual classes)
   - ID matching with globs: `"section#module-*"` uses `filepath.Match`
   - Attribute matching: `"a[href]"`
   - Comma groups: `"h2,h3,h4"`

   For complex deeply nested structures (e.g., `ul > li > ol > li`), the `dq.NewRecursiveNodeMatcher()` supports multi-step patterns with `>` separators and recursive restart. See `pkg/dq/dq.go` for details.

4. **Save HTML fixtures** to `internal/app/testdata/` for golden testing.

### Phase 2: Create the app layer - `internal/app/{name}.go`

Reference files: `internal/app/pydoc.go`, `internal/app/rustdoc.go`, `internal/app/godoc.go`

This file contains all business logic. Follow this structure:

```go
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
```

#### Required public functions (3):

**Search function** - Discovers documents:
```go
func Search{Name}(httpcli *infra.HttpClient, query string) (string, error)
```
- Fetch the search/index page (or use a search API if available)
- Cache the parsed results using the shared `docCache` with key prefix `"{name}:"` and `cache.DefaultExpiration`
- Filter results by case-insensitive substring match
- Return formatted string with results

**Read function** - Paged document reading:
```go
func Read{Name}Paged(httpcli *infra.HttpClient, identifier string, offset, limit int) (string, int, bool, error)
```
- Returns: `(pagedContent, totalLines, hasMore, error)`
- Fetch and parse the document page HTML
- Cache the full parsed content with key `"{name}:{identifier}"`
- Split by `"\n"`, apply offset/limit, return the page

**Search-within function** - Keyword search inside a document:
```go
func SearchWithin{Name}(httpcli *infra.HttpClient, identifier string, keyword string, maxMatches int) (*{Name}SearchResult, error)
```
- Returns a result struct with: `Identifier string`, `Matches []model.SearchMatch`, `Truncated bool`
- Fetch and parse the document (reuses the same cache as Read)
- Use `contentsearch.SearchInContent(content, keyword, maxMatches)` for the search

#### Required parse functions (unexported):

**Document page parser**:
```go
func parse{Name}Page(doc *html.Node) (bool, string)
```
- Uses `dq.Traverse()` with `dq.NewNodeMatcher()` to walk the HTML tree
- Builds a markdown-like text representation
- Returns `(matched, content)` where `matched` indicates the root was found

**Key dq patterns for parsers:**

```go
// Build the output string
var sb strings.Builder

// Root matcher
rootMatcher := dq.NewNodeMatcher(
    dq.NewMatchFunc("section#my-root"),  // Adjust selector
    nil,  // No handler on root itself
    // Child matchers for content inside root:

    // Headings
    dq.NewNodeMatcher(
        dq.NewMatchFunc("h1"),
        func(n *html.Node) {
            sb.WriteString("# " + cleanHeading(dq.InnerText(n, true)) + "\n\n")
        },
    ),

    // Paragraphs
    dq.NewNodeMatcher(
        dq.NewMatchFunc("p"),
        func(n *html.Node) {
            text := strings.TrimSpace(dq.InnerText(n, true))
            if text != "" {
                sb.WriteString(text + "\n")
            }
        },
    ),

    // Code blocks - use RawInnerText to preserve formatting
    dq.NewNodeMatcher(
        dq.NewMatchFunc("pre"),
        func(n *html.Node) {
            sb.WriteString("```\n" + dq.RawInnerText(n, true) + "\n```\n")
        },
    ),

    // Definition lists (API entries)
    dq.NewNodeMatcher(
        dq.NewMatchFunc("dl.myclass"),
        nil,
        // dt = signature, dd = description
        dq.NewNodeMatcher(
            dq.NewMatchFunc("dt"),
            func(n *html.Node) {
                // Use RawInnerText for signatures to avoid extra spaces
                sig := cleanSignature(dq.RawInnerText(n, true))
                sb.WriteString("\n" + sig + "\n")
            },
        ),
        dq.NewNodeMatcher(
            dq.NewMatchFunc("dd"),
            nil,
            // Recurse into dd children (p, pre, nested dl, etc.)
        ),
    ),
)

// Run traversal - pass root matcher directly; dq.Traverse walks children automatically
matched := false
dq.Traverse(doc, []dq.Matcher{rootMatcher})
```

**Important dq notes:**
- `dq.InnerText(n, true)` - Extracts text, adds spaces around element children. Good for prose.
- `dq.RawInnerText(n, true)` - Concatenates raw text nodes. Good for code blocks and signatures.
- Neither strips special characters like `¶` or `[source]` - clean those in helper functions.
- Unmatched matchers carry forward to deeper levels during traversal.
- For nested structures (e.g., methods inside classes), create separate inline matchers to handle one level of nesting.

**URL builder helper**:
```go
func build{Name}URL(identifier string) string {
    return fmt.Sprintf("https://example.com/docs/%s", identifier)
}
```

**Heading/signature cleaners**:
```go
func clean{Name}Heading(text string) string {
    // Remove anchor symbols, collapse spaces, trim
}
```

### Phase 3: Create MCP handlers - `internal/mcptool/{name}.go`

Reference file: `internal/mcptool/pydoc.go`

Define 3 arg structs and 3 handler functions:

```go
package tool

type Search{Name}Args struct {
    Query string `json:"query"`
}

type Read{Name}Args struct {
    Identifier string `json:"identifier_name"`  // Use domain-specific name
    Offset     int    `json:"offset,omitempty"`
    Limit      int    `json:"limit,omitempty"`
}

type SearchWithin{Name}Args struct {
    Identifier string `json:"identifier_name"`
    Keyword    string `json:"keyword"`
    MaxMatches int    `json:"max_matches,omitempty"`
}
```

Each handler function follows this pattern:
1. Validate required args (return `mcp.NewToolResultError()` if missing)
2. Set defaults: `limit` defaults to `app.DefaultLinesPerPage`, `maxMatches` defaults to 10
3. Create `httpcli := infra.NewHttpClient()`
4. Call the app layer function
5. Format and return `mcp.NewToolResultText()`

### Phase 4: Register tools - `internal/mcptool/register.go`

Add 3 tool registrations inside `Register()`, before the final `return nil`:

```go
// search_{name}
tool = mcp.NewTool("search_{name}",
    mcp.WithDescription("Search ... on {site}."),
    mcp.WithString("query", mcp.Required(),
        mcp.Description("Search text to match against ...")),
)
s.AddTool(tool, mcp.NewTypedToolHandler(search{Name}))

// read_{name}
tool = mcp.NewTool("read_{name}",
    mcp.WithDescription("Read ... documentation from {site}. Returns paginated content."),
    mcp.WithString("identifier_name", mcp.Required(),
        mcp.Description("The identifier (e.g., ...)")),
    mcp.WithNumber("offset",
        mcp.DefaultNumber(0),
        mcp.Description("Line offset to start reading from")),
    mcp.WithNumber("limit",
        mcp.DefaultNumber(app.DefaultLinesPerPage),
        mcp.Description("Maximum number of lines to return")),
)
s.AddTool(tool, mcp.NewTypedToolHandler(read{Name}))

// search_within_{name}
tool = mcp.NewTool("search_within_{name}",
    mcp.WithDescription("Search for a keyword within a specific ... page."),
    mcp.WithString("identifier_name", mcp.Required(),
        mcp.Description("The identifier to search within")),
    mcp.WithString("keyword", mcp.Required(),
        mcp.Description("Keyword or phrase to search for")),
    mcp.WithNumber("max_matches",
        mcp.DefaultNumber(10),
        mcp.Description("Maximum number of matches to return")),
)
s.AddTool(tool, mcp.NewTypedToolHandler(searchWithin{Name}))
```

### Phase 5: Create CLI subcommand - `internal/subcmd/{name}.go`

Reference file: `internal/subcmd/pydoc.go`

Create 3 types (note: `search_within` is MCP-only, no CLI subcommand needed):
- `{Name}Cmd` - Main command with `*subcommands.Commander`
- `{Name}SearchCmd` - Search subcommand (no flags, takes query as arg)
- `{Name}ReadCmd` - Read subcommand (flags: `--offset`, `--limit`)

Register the main command in `godevmcp/main.go`:
```go
subcommands.Register(&subcmd.{Name}Cmd{}, "")
```

### Phase 6: Create golden tests - `internal/app/{name}_golden_test.go`

Reference file: `internal/app/pydoc_golden_test.go`

1. Save HTML fixture files to `internal/app/testdata/{name}_*.html`
   - At least 2 document page fixtures (one simple, one complex)
   - One search/index page fixture if applicable

2. Write golden tests:
```go
func TestGolden_parse{Name}Page(t *testing.T) {
    tests := []struct {
        name    string
        fixture string
        golden  string
    }{
        {"simple", "{name}_simple.html", "{name}_simple.golden"},
        {"complex", "{name}_complex.html", "{name}_complex.golden"},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            doc := loadHTMLFixture(t, tt.fixture)
            matched, got := parse{Name}Page(doc)
            require.True(t, matched)
            goldenPath := "testdata/" + tt.golden
            if *update {
                writeGolden(t, goldenPath, got)
                return
            }
            want := readGolden(t, goldenPath)
            assert.Equal(t, want, got)
        })
    }
}
```

3. Generate golden files: `go test -v ./internal/app/ -run TestGolden_parse{Name} -update`
4. Review the generated `.golden` files to verify the output looks correct.

### Phase 7: Verify

Run these commands to verify the implementation:

```bash
make build   # Must compile
make test    # All tests pass (existing + new golden tests)
make lint    # No new issues (pre-existing dupl/goconst warnings are OK)
```

Then test CLI:
```bash
output/godevmcp {name} search <query>
output/godevmcp {name} read <identifier>
```

## Checklist

- [ ] HTML structure explored and documented
- [ ] HTML fixtures saved to `internal/app/testdata/`
- [ ] `internal/app/{name}.go` with Search, ReadPaged, SearchWithin + parsers
- [ ] `internal/mcptool/{name}.go` with 3 arg types + 3 handlers
- [ ] `internal/mcptool/register.go` updated with 3 tool registrations
- [ ] `internal/subcmd/{name}.go` with main + search + read commands (search_within is MCP-only)
- [ ] `godevmcp/main.go` updated with subcommand registration
- [ ] `internal/app/{name}_golden_test.go` with golden tests
- [ ] Golden files generated and reviewed
- [ ] `make build` passes
- [ ] `make test` passes
- [ ] CLI smoke test works
