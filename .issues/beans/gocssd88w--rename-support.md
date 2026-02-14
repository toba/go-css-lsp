---
# gocssd88w
title: Rename support
status: completed
type: feature
priority: low
created_at: 2026-02-01T20:56:16Z
updated_at: 2026-02-01T21:17:36Z
parent: gocsse5ea
---

Rename CSS custom properties across the document.

## Reference
- vscode-css-languageservice: cssNavigation.ts (doRename)

## Checklist
- [ ] Rename custom property definitions and all var() usages
- [ ] Register textDocument/rename and textDocument/prepareRename capabilities
- [ ] Add tests