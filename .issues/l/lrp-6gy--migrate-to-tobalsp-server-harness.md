---
# lrp-6gy
title: Migrate to toba/lsp server harness
status: completed
type: task
priority: high
created_at: 2026-03-21T17:45:43Z
updated_at: 2026-03-21T18:04:39Z
sync:
    github:
        issue_number: "70"
        synced_at: "2026-03-21T18:05:06Z"
---

Replace the hand-rolled LSP main loop with the new `github.com/toba/lsp/server` package (v0.2.0+).

## What changes

The `server` package handles all lifecycle boilerplate:
- JSON-RPC transport via `go.lsp.dev/jsonrpc2`
- `initialize` / `initialized` / `shutdown` / `exit` lifecycle
- Document state management (open/change/close)
- Diagnostic publishing with debouncing
- Optional handler delegation (Hover, Completion, Definition, Formatting, CodeAction, References, Rename, DocumentSymbol)

## Steps

- [x] Add `github.com/toba/lsp v0.2.0` dependency
- [x] Implement `server.Handler` interface (Initialize, Diagnostics, Shutdown)
- [x] Implement any optional handler interfaces (e.g. `server.HoverHandler`, `server.CompletionHandler`)
- [x] Replace main loop with `server.Server{Name: "css-lsp", Version: version, Handler: h}.Run(ctx)`
- [x] Remove hand-rolled JSON-RPC dispatch, document store, and diagnostic goroutine
- [x] Remove direct dependency on `toba/lsp/transport` if no longer needed
- [x] Run tests and linter
- [ ] Verify in editor (VS Code or Zed)

## Summary of Changes
Replaced the entire hand-rolled LSP main loop with `server.Server{}.Run(ctx)`. The new `cssHandler` type implements `server.Handler` (Initialize, Diagnostics, Shutdown) plus all optional handler interfaces: HoverHandler, CompletionHandler, DefinitionHandler, FormattingHandler, CodeActionHandler, ReferencesHandler, RenameHandler, and DocumentSymbolHandler. The server harness manages JSON-RPC transport, document state, and diagnostic publishing with debouncing. Removed direct use of `toba/lsp/transport`. Updated dependency to `toba/lsp v0.2.1`.
