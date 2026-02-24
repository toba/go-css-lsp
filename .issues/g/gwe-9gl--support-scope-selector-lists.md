---
# gwe-9gl
title: Support @scope selector lists
status: completed
type: feature
priority: normal
tags:
    - upstream
created_at: 2026-02-24T17:51:51Z
updated_at: 2026-02-24T18:03:17Z
---

Upstream vscode-css-languageservice added support for selector lists in `@scope` rules ([39d6045](https://github.com/microsoft/vscode-css-languageservice/commit/39d6045682c5a10ca82d81b1b4ce4ef5de97f2fa), PR #474).

Ensure our parser handles `@scope` with selector lists correctly.

- [x] Review upstream changes to `cssParser.ts`
- [x] Add failing test for `@scope` with selector lists
- [x] Implement parser support (already handled by generic at-rule parsing)
- [x] Run tests


## Summary of Changes

Our parser already handles `@scope` with selector lists correctly because it treats `@scope` as a generic block at-rule, collecting all prelude tokens (including commas for selector lists) without specific validation. This differs from the upstream vscode-css-languageservice which had a specialized `_parseSelector(true)` call that rejected comma-separated selectors.

Added test coverage:
- `TestParseScopeRule` in `internal/css/parser/parser_test.go` — 6 subtests covering single root, selector lists, root+limit, both with selector lists, implicit scope, and limit-only
- `TestFormat_ScopeWithSelectorList` in `internal/css/analyzer/format_test.go` — 4 subtests verifying formatting preserves selector lists in `@scope` preludes

No parser changes were needed.
