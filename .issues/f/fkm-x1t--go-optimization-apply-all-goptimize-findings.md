---
# fkm-x1t
title: 'Go optimization: apply all goptimize findings'
status: completed
type: task
priority: normal
created_at: 2026-02-16T21:37:34Z
updated_at: 2026-02-16T21:44:52Z
---

Apply all 22 findings from goptimize report:
- [ ] Remove redundant golangci-lint linters (intrange, modernize, copyloopvar)
- [ ] Modern idioms: wg.Go, log.Printf→slog, use constants for code action kinds
- [ ] Function extraction: unmarshalRequest, resolveDocument, marshalResponse in main.go
- [ ] Function extraction: unify writeValue/writeSingleSelector pairs in format.go
- [ ] Function extraction: completion tagging helper, close-paren helper
- [ ] Generics: lookup[T] in data.go, labToLCH in color.go
- [ ] Constants: panicStackBufSize, important length, markdown kind
- [ ] Concurrency: errgroup in ScanWorkspace
- [ ] Test coverage gaps

## Summary of Changes

**9 files changed, 221 insertions, 488 deletions (-267 net lines)**

### Applied
- Removed redundant `copyloopvar` linter (no-op on Go 1.22+)
- Migrated `log.Printf` → `slog.Warn` in `parsing.go`
- Replaced hardcoded code action kind strings with constants from `analyzer` package
- Replaced `wg.Add/Done` with `wg.Go` (Go 1.26) in `parsing_test.go`
- Extracted `unmarshalRequest[T]`, `marshalResponse`, `resolveDocument` helpers in `main.go` (eliminated ~190 lines across 44 call sites)
- Extracted generic `lookup[T]` helper in `data.go` (5 Lookup functions)
- Extracted `labToLCH` helper in `color.go` (shared by `colorToLCH` and `colorToOklch`)
- Unified `writeSingleSelector`/`writeSingleSelectorTo` and `writeValue`/`writeValueTo` in `format.go`
- Extracted `tagCompletionItem` helper in `completion.go` (4 call sites)
- Refactored inline close-paren loops in `parseOriginColor` to reuse `skipPastCloseParen`
- Added constants: `panicStackBufSize`, `markupKindMarkdown`
- Replaced magic number `11` with `len(" !important")`

### Skipped
- Concurrency (errgroup in ScanWorkspace): would add first external dependency (`golang.org/x/sync`) to a zero-dependency project
- Test coverage gaps: deferred to separate issue
