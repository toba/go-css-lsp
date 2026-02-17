---
# gocssqegm
title: Selection ranges
status: completed
type: feature
priority: low
created_at: 2026-02-01T20:56:16Z
updated_at: 2026-02-01T21:17:36Z
parent: gocsse5ea
sync:
    github:
        issue_number: "52"
        synced_at: "2026-02-17T18:03:22Z"
---

Provide smart selection ranges — expand/shrink selection by semantic CSS units.

## Reference
- vscode-css-languageservice: cssSelectionRange.ts

## Checklist
- [ ] Selection expands: value → declaration → rule block → ruleset → stylesheet
- [ ] Register textDocument/selectionRange capability
- [ ] Add tests
