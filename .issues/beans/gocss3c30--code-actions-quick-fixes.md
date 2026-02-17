---
# gocss3c30
title: Code actions / quick fixes
status: completed
type: feature
priority: normal
created_at: 2026-02-01T20:56:16Z
updated_at: 2026-02-01T21:09:34Z
parent: gocsse5ea
sync:
    github:
        issue_number: "62"
        synced_at: "2026-02-17T18:03:22Z"
---

Provide quick fixes for diagnostics — e.g. suggest similar property names for unknown properties.

## Reference
- vscode-css-languageservice: cssCodeActions.ts

## Checklist
- [ ] Suggest similar property names for unknown property diagnostics
- [ ] Register textDocument/codeAction capability
- [ ] Add tests
