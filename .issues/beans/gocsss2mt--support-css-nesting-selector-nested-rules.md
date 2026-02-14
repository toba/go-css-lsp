---
# gocsss2mt
title: Support CSS nesting (& selector, nested rules)
status: completed
type: bug
priority: normal
created_at: 2026-02-01T22:20:15Z
updated_at: 2026-02-01T22:35:13Z
---

## Problem

The parser does not support CSS Nesting (spec: CSS Nesting Module Level 1). This causes two issues:

1. **Formatting destroys nested CSS** — the formatter produces garbled output because `parseRuleset()` only parses declarations in a ruleset body, never nested rulesets. Nested selectors, `&` combinators, and nested blocks are mangled.

2. **False diagnostics everywhere** — unknown property errors fire on nested selectors since they are parsed as declarations.

## Root Cause

- `parseRuleset()` (parser.go:240) only calls `parseDeclaration()` in its body loop
- `parseBlock()` (parser.go:147) already handles disambiguation between declarations and nested rulesets via `looksLikeDeclaration()`, but this logic is only used for at-rule blocks (e.g. `@media`)
- The `Ruleset` AST node only has `Declarations []*Declaration`, no way to store nested children

## Checklist

- [x] Modify `Ruleset` AST node to support nested children (`Children []Node` alongside or replacing `Declarations`)
- [x] Update `parseRuleset()` body loop to use `looksLikeDeclaration()` disambiguation (like `parseBlock()` does)
- [x] Handle `&` nesting selector in `parseSelector()` (parsed as Delim token, works naturally)
- [x] Update `Walk()` in ast_walk.go to traverse nested children
- [x] Update formatter (format.go) to recursively format nested rulesets inside rulesets
- [x] Update diagnostics to not flag nested selectors as unknown properties
- [x] Update all other analyzers that walk rulesets (symbols, folding, selection, highlights, references, definition, rename, completion, hover, links, code actions, color)
- [x] Add parser tests for nested CSS
- [x] Add formatter tests for nested CSS
- [x] Add diagnostics tests for nested CSS
- [x] Run `golangci-lint run --fix` and `go test ./...`