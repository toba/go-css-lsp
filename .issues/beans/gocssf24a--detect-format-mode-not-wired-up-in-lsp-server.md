---
# gocssf24a
title: Detect format mode not wired up in LSP server
status: completed
type: bug
priority: normal
created_at: 2026-02-01T23:16:43Z
updated_at: 2026-02-01T23:19:18Z
sync:
    github:
        issue_number: "16"
        synced_at: "2026-02-17T18:03:20Z"
---

## Problem

The `"detect"` format mode was never handled in the LSP server switch statement in `cmd/go-css-lsp/main.go`. When a user configured `"formatMode": "detect"` in initializationOptions, the switch fell through to the default case (expanded mode), making detect mode completely non-functional via LSP.

Additionally, the README did not document the detect mode.

## Root Cause

`main.go:785-790` had cases for `"compact"` and `"preserve"` but not `"detect"`. The internal formatter dispatch (`format.go:127`) and all detect logic (`formatRulesetDetect`) were correct — they just never got invoked through the LSP path.

## Checklist

- [x] Add `"detect"` to README formatting modes table, settings table, and features summary
- [x] Add `case "detect"` to format mode switch in `cmd/go-css-lsp/main.go`
- [x] Add test for multi-selector detect mode (e.g., `ul, ol { ... }`)
- [x] Verify end-to-end with `go test ./...` and `golangci-lint run`
