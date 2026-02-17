---
# gocssp73x
title: Fix concurrent map access crash in LSP server
status: completed
type: bug
priority: critical
created_at: 2026-02-02T04:54:28Z
updated_at: 2026-02-02T04:56:20Z
sync:
    github:
        issue_number: "26"
        synced_at: "2026-02-17T18:03:20Z"
---

storage.RawFiles and storage.ParsedFiles are plain Go maps shared between the diagnostic goroutine (processDiagnosticNotification, writes at lines 1600/1603) and all request handlers (reads at 15+ locations). Go's runtime panics with an unrecoverable fatal error on concurrent map read/write, which cannot be caught by defer/recover. This kills the server process.

## Fix
Add a sync.RWMutex to workspaceStore. The diagnostic goroutine takes a write lock when updating files. Request handlers take a read lock when accessing files.

## Checklist
- [x] Add sync.RWMutex field to workspaceStore
- [x] Lock for write in processDiagnosticNotification when updating RawFiles/ParsedFiles
- [x] Lock for read in all request handlers accessing RawFiles/ParsedFiles
- [x] Also add panic recovery to processDiagnosticNotification goroutine
- [x] Run tests with -race flag
- [x] Run golangci-lint
