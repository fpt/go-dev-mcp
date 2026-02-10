# dq Package Improvement Plan

The `pkg/dq` package provides HTML document querying for the doc parsers (godoc, rustdoc, pydoc). It was originally built with minimum functionality for godoc parsing. After adding rustdoc and pydoc support, several recurring workarounds have emerged.

## Current Workarounds in Parsers

### Manual node walking replaces missing `FindOne`/`FindAll`
- `godoc.go:findVersionInNode()` — manually recurses FirstChild/NextSibling to locate a `<strong>` element
- `pydoc.go:findDirectChild()` — walks siblings to find a child by tag name
- `pydoc.go:extractModuleFromRow()` — manually iterates `<td>` children by index

### Manual attribute value checking replaces missing `[attr=val]` selectors
- `pydoc.go:252` — `dq.GetHref(a)` then `strings.Contains(href, "#module-")` to filter links
- `rustdoc.go:212` — captures `dq.GetHref(n)` inside handler to read href value

### Hardcoded godoc logic leaks into core library
- `dq.go:InnerText()` skips `span.Documentation-sinceVersion` — a pkg.go.dev-specific class baked into the generic dq package

### `HasChild` called inside handlers instead of in matchers
- `godoc.go:289` — `dq.HasChild(n, "a")` inside an `<li>` handler to conditionally process list items

### `GetHref` is too narrow
- Only extracts `href`. Other attributes (`src`, `data-*`, `role`, `aria-*`) require manual `n.Attr` iteration.

## Planned Improvements

### High Priority (Done)

#### 1. Attribute value matching in selectors

Extend `matchSingle` to support CSS attribute value selectors:

| Syntax | Meaning |
|--------|---------|
| `[attr=val]` | Exact match |
| `[attr*=val]` | Contains substring |
| `[attr^=val]` | Starts with prefix |
| `[attr$=val]` | Ends with suffix |

Examples:
```
a[href*="#module-"]     — links containing "#module-" in href
div[role=note]          — div elements with role="note"
a[href^="/3/library/"]  — links starting with "/3/library/"
```

Combined with existing selectors: `a.release[href^="/crate/"]`

**Files:** `pkg/dq/dq.go` (`matchSingle`, new `matchAttrExpr`, `matchAttrValue`)

#### 2. `FindOne`, `FindAll`, and `FindDirectChild`

Add query functions that return nodes instead of requiring handler callbacks:

```go
func FindOne(root *html.Node, selector string) *html.Node
func FindAll(root *html.Node, selector string) []*html.Node
func FindDirectChild(n *html.Node, selector string) *html.Node
```

`FindOne`/`FindAll` search the entire subtree depth-first. `FindDirectChild` checks only immediate children.

Replaces: `findVersionInNode`, `findDirectChild`, manual `<td>` iteration.

**Parsers refactored:**
- `godoc.go:findVersionInNode` — now uses `dq.FindAll(n, "strong")`
- `pydoc.go:extractModuleFromRow` — now uses `dq.FindAll(tr, "td")` + `dq.FindDirectChild(td, "a[href*=#module-]")`
- `pydoc.go:findDirectChild` — removed, replaced by `dq.FindDirectChild`

**Files:** `pkg/dq/dq.go`

### Medium Priority

#### 3. `GetAttr` generic helper (Done)

```go
func GetAttr(n *html.Node, key string) string
```

Generic attribute getter. `GetHref` becomes `GetAttr(n, "href")` (keep `GetHref` as alias for convenience).

**Files:** `pkg/dq/dq.go`

#### 4. Remove hardcoded godoc filters from `InnerText` (Done)

Added `InnerTextWithFilter(n, recurse, filter)` where `NodeFilter` returns true to include a node.
`InnerText` now delegates to `InnerTextWithFilter` with `DefaultNodeFilter` (skips `a[aria-label]` only).
The godoc-specific `span.Documentation-sinceVersion` skip moved to `godocNodeFilter` in `godoc.go`.
`parseDocument` and `parseReadme` use a local `innerText` closure wrapping the godoc filter.

**Files:** `pkg/dq/dq.go`, `internal/app/godoc.go`

### Low Priority

#### 5. `:has()` compound matcher

A match function combinator: `li:has(a)` matches `<li>` only if it contains a direct child `<a>`. This replaces `dq.HasChild()` calls inside handlers.

This is lower priority because `HasChild` already works — it's just not as declarative.

**Files:** `pkg/dq/dq.go` (`matchSingle`)
