---
name: verify-doc-parsing
description: |
  This skill should be used when the user asks to verify, audit, or review that
  the dq-based documentation parsers still work against the live site. Trigger
  phrases include "verify doc parsing", "check pkg.go.dev parsing still works",
  "is our godoc/rustdoc/pydoc parser still correct", "check for doc site drift",
  "review our dq-based parsing", "browse docs.rs and review parsing",
  "did pkg.go.dev change its HTML". It uses the browser-sandbox MCP tool to
  browse the live documentation site and compare the real DOM against the CSS
  selectors the parser depends on, flagging drift and robustness gaps.
---

# Verify dq-based Documentation Parsing

The go-dev-mcp server scrapes documentation sites (pkg.go.dev, docs.rs,
docs.python.org) by walking HTML with the `pkg/dq` package and matching on
specific CSS classes/IDs. Those selectors are **brittle**: if the site changes
its markup, the parser silently returns empty output and the golden tests
(which run against *saved* HTML) won't notice. This skill verifies the parsers
against the **live** site.

## When to use which parser file

| Site | Parser | Key selectors to confirm |
|------|--------|--------------------------|
| pkg.go.dev | `internal/app/godoc.go` | `div.Documentation`, `section.Documentation-*`, `div.Documentation-declaration > pre`, `span.Documentation-sinceVersion`, `div.SearchResults(-summary)`, `div.SearchSnippet(-infoLabel)`, `div.Overview-readmeContent` |
| docs.rs | `internal/app/rustdoc.go` | confirm from the file's `dq.NewMatchFunc(...)` calls |
| docs.python.org | `internal/app/pydoc.go` | confirm from the file's `dq.NewMatchFunc(...)` calls |

Always derive the selector list from the code, not memory — `grep` it:

```bash
grep -n "NewMatchFunc\|hasClass\|NodeFilter\|Overview-readme\|Documentation-" internal/app/<parser>.go
```

## Step-by-step

### Phase 1: Extract what the parser expects

1. Read the target parser (`internal/app/{godoc,rustdoc,pydoc}.go`). List every
   selector passed to `dq.NewMatchFunc(...)` and every class checked in custom
   `NodeFilter` functions. These are your verification targets.
2. Note the **page types** the parser handles. For pkg.go.dev there are three:
   - **doc page** (`parseDocument`) — e.g. `https://pkg.go.dev/golang.org/x/net/html`
   - **search page** (`parseSearchResult`) — e.g. `https://pkg.go.dev/search?q=yaml`
   - **README page** (`parseReadme`) — e.g. `https://pkg.go.dev/github.com/stretchr/testify`
3. Recall the dq traversal rule (see `pkg/dq/dq.go` `Traverse`): an **unmatched
   matcher propagates down to descendants**. So a `pre` matcher placed under a
   `div` matcher still reaches a `<pre>` nested several levels deeper (e.g. inside
   `div.Documentation-declaration`). Don't flag deep nesting as broken — verify
   the selector exists *somewhere* under the expected root.

### Phase 2: Browse the live site (browser-sandbox MCP)

Start and always end with instance lifecycle:

```
mcp__browser-sandbox__create-chrome-instance (headless: true)  -> returns instance id
... work ...
mcp__browser-sandbox__close (id)
```

For each page type:

1. **Navigate scoped to the parser's root** to avoid serializing the whole
   ad/nav-heavy page (which truncates or times out). Use the `selector` arg:

   ```
   mcp__browser-sandbox__navigate(id, url, selector=".Documentation", depth=2)
   ```

2. **Confirm each expected selector is present** in one call with multi-extract.
   An empty list means the selector matched nothing → likely drift:

   ```
   mcp__browser-sandbox__multi-extract(id, selectors=[
     {name:"documentation",   selector:".Documentation"},
     {name:"declaration",     selector:".Documentation-declaration"},
     {name:"sinceVersion",    selector:".Documentation-sinceVersion"},
     {name:"readmeContent",   selector:".Overview-readmeContent"},
   ])
   ```

3. **Inspect internal structure** of one representative subtree with
   `get-all-elements(selector=..., depth=5)` to confirm nesting (e.g. that
   signatures still live in `div.Documentation-function > div.Documentation-declaration > pre`).

### Phase 3: CRITICAL gotcha — text nodes vanish at depth limits

`navigate` and `get-all-elements` return a **depth-limited DOM tree that drops
bare text nodes**. This will make you see `<h2><span>(path)</span></h2>` and
wrongly conclude the package name was removed, when the real `<h2>` is
`yaml (gopkg.in/yaml.v3)` (a text node + span).

**Always confirm actual rendered text with `multi-extract`'s `text` field**
(it returns full `textContent`), not the structural tree:

```
mcp__browser-sandbox__multi-extract(id, selectors=[
  {name:"snippetTitle", selector:"//a[@data-test-id='snippet-title']"},
  {name:"summary",      selector:".SearchResults-summary"},
])
```

If a structural observation suggests content is missing, **re-verify with a text
extract before reporting it as a regression.** (This exact trap produced two
false positives in a prior review.)

### Phase 4: Distinguish live drift from existing behavior

A selector being absent live is only a problem if the golden fixture relied on
it. Cross-check:

```bash
ls internal/app/testdata/                      # *.html fixtures + *.golden expected output
cat internal/app/testdata/<parser>_*.golden    # what the parser is expected to emit
```

If live structure == golden fixture structure, there's no drift. If they differ,
the saved fixture is stale and the live output has degraded — that's a real
finding the tests can't catch.

### Phase 5: Robustness review (independent of drift)

Check these failure-handling patterns in each parser (they tend to be copied
across godoc/rustdoc/pydoc):

1. **Silent empty on total parse failure.** When all `parse*` functions return
   `matched == false`, does the public function return `""` with `nil` error, or
   a real error (`ErrNotFound`)? Returning empty-as-success hides drift.
   Grep: `grep -n "matched\|_, parsed\|Cache.Set\|ErrNotFound" internal/app/<parser>.go`
2. **Caching failures.** Is `docCache.Set(...)` called even when the parsed
   document is empty? That poisons the cache for the TTL (30 min) on a transient
   miss. Only cache non-empty successful parses.

## Reporting

Summarize as:
- **Healthy**: selectors confirmed present live (list them).
- **Drift**: selectors/structure that changed vs. the parser/golden, with the
  live evidence (the multi-extract text, not the truncated tree).
- **Robustness**: silent-empty / cache-poisoning gaps with file:line refs.
- **False alarms ruled out**: anything that looked broken in the structural tree
  but was confirmed fine via text extract.

Propose concrete fixes; offer to implement on a branch. Do not edit parsers as
part of verification unless asked.

## Checklist

- [ ] Selector list derived from the parser source (not memory)
- [ ] Chrome instance created and **closed** at the end
- [ ] All page types browsed (doc / search / README for pkg.go.dev)
- [ ] Each expected selector confirmed via multi-extract (empty list = drift)
- [ ] Suspected "missing content" re-verified with a text extract (Phase 3)
- [ ] Findings cross-checked against golden fixtures (drift vs. existing)
- [ ] Robustness patterns (silent-empty, cache-on-failure) reviewed
- [ ] Findings reported with live evidence + concrete fix proposals
