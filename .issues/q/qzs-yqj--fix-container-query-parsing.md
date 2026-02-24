---
# qzs-yqj
title: Fix @container query parsing
status: completed
type: bug
priority: normal
tags:
    - upstream
created_at: 2026-02-24T17:51:51Z
updated_at: 2026-02-24T17:56:14Z
sync:
    github:
        issue_number: "66"
        synced_at: "2026-02-24T18:07:11Z"
---

Upstream vscode-css-languageservice fixed `@container` query parsing ([54c68ce](https://github.com/microsoft/vscode-css-languageservice/commit/54c68cec52cf7cf0aac0ad45297c5b22e17d73ce), PR #473).

Check whether our parser has the same issue and port the fix if needed.

- [ ] Review upstream changes to `cssParser.ts`
- [ ] Write failing test reproducing the issue
- [ ] Port fix
- [ ] Run tests


## Summary of Changes

Investigation found that our parser is **not affected** by the upstream bug. The upstream fix ([54c68ce](https://github.com/microsoft/vscode-css-languageservice/commit/54c68cec52cf7cf0aac0ad45297c5b22e17d73ce)) addressed three issues in their structured `@container` parser:

1. Container name without query: `@container card { }`
2. Comma-separated queries: `@container (inline-size > 30em), style(--responsive: true) { }`
3. Standalone custom properties in `style()`: `@container style(--responsive) { }`

Our parser uses a generic `parseAtRule()` that collects prelude tokens without interpreting container query structure, so all three cases already parse correctly.

Added `TestParseContainerRule` with 8 subtests covering all upstream test cases to prevent regressions.
