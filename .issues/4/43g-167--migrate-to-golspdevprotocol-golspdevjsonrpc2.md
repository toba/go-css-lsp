---
# 43g-167
title: Migrate to go.lsp.dev/protocol + go.lsp.dev/jsonrpc2
status: completed
type: feature
priority: normal
created_at: 2026-03-21T17:10:58Z
updated_at: 2026-03-21T18:04:39Z
sync:
    github:
        issue_number: "69"
        synced_at: "2026-03-21T18:05:06Z"
---

Replace the hand-rolled LSP protocol types and JSON-RPC transport with the standard go.lsp.dev/protocol and go.lsp.dev/jsonrpc2 packages. This eliminates maintaining custom LSP struct definitions and transport code, and gives full spec-compliant types for all LSP methods.

## Steps
- [x] Add go.lsp.dev/protocol and go.lsp.dev/jsonrpc2 as dependencies
- [x] Replace all custom LSP protocol types with imports from go.lsp.dev/protocol
- [x] Replace the custom JSON-RPC transport with go.lsp.dev/jsonrpc2 stream handling
- [x] Update all handler functions to use the standard protocol types
- [x] Remove custom type definitions that are now redundant
- [x] Run tests and linter

## Summary of Changes
Upgraded toba/lsp to v0.2.1 which brings in go.lsp.dev/protocol and go.lsp.dev/jsonrpc2. Rewrote main.go to use protocol types (protocol.Range, protocol.Position, protocol.Hover, etc.) instead of the hand-rolled lsp package types. JSON-RPC transport is now handled by the toba/lsp/server harness via go.lsp.dev/jsonrpc2. The old custom lsp package types are no longer imported by main.
