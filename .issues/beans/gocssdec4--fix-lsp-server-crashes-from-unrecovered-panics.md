---
# gocssdec4
title: Fix LSP server crashes from unrecovered panics
status: completed
type: bug
priority: normal
created_at: 2026-02-02T04:20:56Z
updated_at: 2026-02-02T04:25:20Z
sync:
    github:
        issue_number: "39"
        synced_at: "2026-02-17T18:03:20Z"
---

The LSP server crashes when any panic occurs during request handling because:
1. No panic recovery in the main request loop
2. Explicit panic() calls in notification handlers (didOpen, didChange, didClose)
3. ID.UnmarshalJSON doesn't bounds-check empty data

## Checklist
- [ ] Add panic recovery to the main request loop that logs the panic and returns an error response
- [ ] Replace panic() calls in ProcessDidOpenTextDocumentNotification with error returns
- [ ] Replace panic() calls in ProcessDidChangeTextDocumentNotification with error returns
- [ ] Replace panic() calls in ProcessDidCloseTextDocumentNotification with error returns
- [ ] Replace panic() calls in ProcessInitializeRequest with error returns
- [ ] Replace panic() calls in ProcessShutdownRequest with error returns
- [ ] Replace panic() calls in ProcessIllegalRequestAfterShutdown with error returns
- [ ] Add bounds check to ID.UnmarshalJSON for empty data
- [ ] Add test for ID.UnmarshalJSON with empty data
- [ ] Run tests and linter
