---
# gocsse5ea
title: CSS3 Language Server — Feature Parity
status: completed
type: milestone
priority: normal
created_at: 2026-02-01T20:55:41Z
updated_at: 2026-02-01T21:19:22Z
---

Bring the Go CSS LSP to feature parity with vscode-css-languageservice (CSS3 only, no SCSS/Sass/LESS). Reference projects: microsoft/vscode-css-languageservice and lmn451/css-variables-zed.

## Already Implemented
- Tokenizer/scanner (CSS3)
- Recursive descent parser with error recovery
- Diagnostics: unknown properties, duplicates, empty rulesets, unknown at-rules
- Hover: properties, at-rules, functions, pseudo-classes, pseudo-elements
- Completion: properties, values, at-rules, pseudo-classes/elements, selectors, functions
- LSP lifecycle: init, didOpen/didChange/didClose, hover, completion, publishDiagnostics, shutdown/exit
- Property database (~450+ properties, 14 at-rules, 30+ pseudo-classes, 147 named colors)