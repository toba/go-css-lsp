# go-css-lsp

A Language Server Protocol implementation for CSS3, written in Go.

CSS3 only — no SCSS, Sass, or LESS support.

## Features

- **Diagnostics** — unknown properties, duplicate properties, unknown at-rules, empty rulesets, zero-with-unit hints, `!important` hints, vendor prefix hints, parse errors
- **Hover** — property documentation on hover
- **Completion** — properties, values, at-rules, pseudo-classes, pseudo-elements, HTML elements, color functions
- **Document Colors** — color picker integration for named colors, hex, `rgb()`, `hsl()`, `hwb()`
- **Color Presentations** — convert between color formats
- **Document Symbols** — outline view for rulesets, at-rules, custom properties
- **Go to Definition** — jump to CSS custom property declarations
- **Find References** — find all usages of variables and selectors
- **Rename** — rename CSS custom properties across the document
- **Document Highlights** — highlight all occurrences of a symbol
- **Code Actions** — quick fixes for diagnostics
- **Folding Ranges** — code folding for rulesets and at-rules
- **Document Links** — clickable URLs in `@import` and `url()`
- **Document Formatting** — CSS pretty-printer
- **Selection Ranges** — expand/shrink selection
- **Workspace Variable Indexing** — index CSS custom properties across workspace files

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

## Acknowledgments

This project was built using the following projects as references:

- [microsoft/vscode-css-languageservice](https://github.com/microsoft/vscode-css-languageservice) (MIT License) — feature reference for LSP capabilities and CSS language intelligence
- [lmn451/css-variables-zed](https://github.com/lmn451/css-variables-zed) (GPL-3.0 License) — reference for CSS variable workspace indexing approach

## License

MIT License — see [LICENSE.md](LICENSE.md) for details.
