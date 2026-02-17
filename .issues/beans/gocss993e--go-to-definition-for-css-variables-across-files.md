---
# gocss993e
title: Go to Definition for CSS variables across files
status: completed
type: bug
priority: normal
created_at: 2026-02-01T23:21:02Z
updated_at: 2026-02-01T23:21:49Z
sync:
    github:
        issue_number: "2"
        synced_at: "2026-02-17T18:03:20Z"
---

processDefinition only searches the current file's stylesheet. It never consults the workspace VarIndex. When a custom property is defined in a different CSS file, the LSP returns null.

## Checklist
- [x] Add VarReferenceAt to internal/css/css.go
- [x] Update processDefinition in main.go to fall back to workspace index
- [x] Add test for cross-file definition lookup
- [x] Run tests and linter
