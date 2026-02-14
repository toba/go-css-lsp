---
# gocss5h3a
title: Document colors and color presentations
status: completed
type: feature
priority: high
created_at: 2026-02-01T20:56:16Z
updated_at: 2026-02-01T21:03:51Z
parent: gocsse5ea
---

Detect color values in CSS documents and provide color presentations (format conversions).

## Reference
- vscode-css-languageservice: cssNavigation.ts (findDocumentColors, getColorPresentations)

## Checklist
- [ ] Detect hex colors (#rgb, #rrggbb, #rrggbbaa)
- [ ] Detect named colors (red, blue, etc.)
- [ ] Detect rgb()/rgba() function colors
- [ ] Detect hsl()/hsla() function colors
- [ ] Detect hwb(), lab(), lch(), oklch(), oklab() colors
- [ ] Return DocumentColor responses with LSP color range
- [ ] Implement color presentations (convert between formats)
- [ ] Register textDocument/documentColor capability
- [ ] Add tests