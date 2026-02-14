---
# gocssnrkb
title: Find references
status: completed
type: feature
priority: normal
created_at: 2026-02-01T20:56:16Z
updated_at: 2026-02-01T21:08:06Z
parent: gocsse5ea
blocking:
    - gocsski7q
    - gocssd88w
---

Find all references to a CSS custom property within the document.

## Reference
- vscode-css-languageservice: cssNavigation.ts (findReferences)

## Checklist
- [ ] Find all usages of a custom property (--var-name and var(--var-name))
- [ ] Register textDocument/references capability
- [ ] Add tests
- [ ] (Future) Workspace-wide references