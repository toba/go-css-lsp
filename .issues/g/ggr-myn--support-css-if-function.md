---
# ggr-myn
title: Support CSS if() function
status: completed
type: feature
priority: normal
tags:
    - upstream
created_at: 2026-02-24T17:51:51Z
updated_at: 2026-02-24T17:59:31Z
---

Upstream vscode-css-languageservice added support for the new CSS `if()` function ([2a8cf1f](https://github.com/microsoft/vscode-css-languageservice/commit/2a8cf1f48c24acef49206826aabab1da85eab2ab), PR #472).

Add parsing and completion support for `if()`.

- [x] Review upstream changes and CSS spec for `if()`
- [x] Add failing test for `if()` usage
- [x] Implement parser support
- [x] Add completions if applicable
- [x] Run tests


## Summary of Changes

The parser already handles `if()` correctly — our generic `parseValue()` tracks paren depth, so semicolons inside `if()` do not break value parsing. No parser changes needed.

Added `if()` to the LSP data layer:
- `internal/css/data/values.go`: Added `"if"` to `CommonFunctions` (enables completions)
- `internal/css/data/functions.go`: Added `Function` entry with description, MDN link, and signatures (enables hover)
- `internal/css/parser/parser_test.go`: Added `TestParseIfFunction` with 6 subtests covering various `if()` syntaxes
