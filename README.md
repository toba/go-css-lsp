# go-css-lsp

A Language Server Protocol implementation for CSS3, written in Go. Designed as a lightweight, fast alternative to [vscode-css-languageservice](https://github.com/microsoft/vscode-css-languageservice).

CSS3 only — no SCSS, Sass, or LESS.

## Features

| Category | Capabilities |
|----------|-------------|
| **Diagnostics** | Unknown properties, duplicates, unknown at-rules, empty rulesets, `!important` hints, vendor prefix hints, zero-with-unit hints, parse errors |
| **Hover** | Property documentation with MDN references |
| **Completion** | Properties, values, at-rules, pseudo-classes, pseudo-elements, HTML elements, color functions |
| **Colors** | Color picker for hex, named colors, `rgb()`, `hsl()`, `hwb()`, `lab()`, `lch()`, `oklab()`, `oklch()`; convert between formats |
| **Navigation** | Go to definition, find references, document symbols, document highlights |
| **Editing** | Rename CSS custom properties, code actions (quick fixes), formatting (supports format-on-save), selection ranges |
| **Structure** | Folding ranges, document links (`@import`, `url()`) |
| **Workspace** | Cross-file CSS custom property indexing |

## Editor Support

Used by the [Gossamer](https://github.com/toba/gossamer) Zed extension. Compatible with any editor that supports LSP.

## Installation

Download a binary from [Releases](https://github.com/toba/go-css-lsp/releases), or build from source:

```bash
go install github.com/toba/go-css-lsp/cmd/go-css-lsp@latest
```

## Building

```bash
go build ./cmd/go-css-lsp
```

## Testing

```bash
go test ./...
```

## Credits

This project was built using the following as references for feature design and architecture:

- **[microsoft/vscode-css-languageservice](https://github.com/microsoft/vscode-css-languageservice)** (MIT) — CSS language intelligence and LSP feature set. The primary reference for what capabilities a CSS language server should provide.
- **[lmn451/css-variables-zed](https://github.com/lmn451/css-variables-zed)** (GPL-3.0) — workspace-wide CSS variable indexing approach.

No code was copied from either project. They served as feature and architectural references for an independent Go implementation.

## License

MIT — see [LICENSE](LICENSE) for details.
