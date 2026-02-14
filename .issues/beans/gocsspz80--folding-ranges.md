---
# gocsspz80
title: Folding ranges
status: completed
type: feature
priority: normal
created_at: 2026-02-01T20:56:16Z
updated_at: 2026-02-01T21:11:39Z
parent: gocsse5ea
---

Provide folding ranges for CSS rule blocks, at-rules, and comment blocks.

## Reference
- vscode-css-languageservice: cssFolding.ts

## Checklist
- [ ] Fold rule blocks ({ ... })
- [ ] Fold at-rule blocks (@media, @keyframes)
- [ ] Fold multi-line comments
- [ ] Register textDocument/foldingRange capability
- [ ] Add tests