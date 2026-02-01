---
# gocssz65p
title: Document symbols
status: completed
type: feature
priority: high
created_at: 2026-02-01T20:56:16Z
updated_at: 2026-02-01T21:05:19Z
parent: gocsse5ea
---

Provide document symbol outline for CSS files — selectors, at-rules, custom properties.

## Reference
- vscode-css-languageservice: cssNavigation.ts (findDocumentSymbols, findDocumentSymbols2)

## Checklist
- [ ] Extract selector symbols from rulesets
- [ ] Extract at-rule symbols (@media, @keyframes, etc.)
- [ ] Extract custom property definitions (--var-name)
- [ ] Support DocumentSymbol hierarchy (nested rules)
- [ ] Register textDocument/documentSymbol capability
- [ ] Add tests