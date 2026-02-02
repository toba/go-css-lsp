# go-css-lsp

A Language Server Protocol implementation for CSS3, written in Go. Designed as a lightweight, fast alternative to [vscode-css-languageservice](https://github.com/microsoft/vscode-css-languageservice).

CSS3 only — no SCSS, Sass, or LESS.

## Features

| Category | Capabilities |
|----------|-------------|
| **Diagnostics** | Unknown properties, duplicates, unknown at-rules, experimental property warnings, deprecated property warnings, empty rulesets, `!important` hints, vendor prefix hints, zero-with-unit hints, parse errors |
| **Hover** | Property documentation with MDN references, experimental status indicators |
| **Completion** | Properties, values, at-rules, pseudo-classes, pseudo-elements, HTML elements, color functions; experimental features tagged |
| **Colors** | Color picker for hex, named colors, `rgb()`, `hsl()`, `hwb()`, `lab()`, `lch()`, `oklab()`, `oklch()`; convert between formats |
| **Navigation** | Go to definition, find references, document symbols, document highlights |
| **Editing** | Rename CSS custom properties, code actions (quick fixes), formatting (expanded/compact/preserve/detect modes), selection ranges |
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

## Formatting Modes

The formatter supports four modes, configured via `initializationOptions`:

| Mode | Behavior |
|------|----------|
| **expanded** (default) | One declaration per line |
| **compact** | Single-line rulesets when they fit within `printWidth` |
| **preserve** | Keeps original single/multi-line layout, normalizes whitespace |
| **detect** | Infers intent from source: if the first property is on the same line as `{` and the result fits `printWidth`, single-line; otherwise expanded |

Rulesets containing nested rules always use expanded format regardless of mode.

```json
{
  "initializationOptions": {
    "formatMode": "compact",
    "printWidth": 80,
    "experimentalFeatures": "warning",
    "deprecatedFeatures": "warning"
  }
}
```

| Setting | Type | Default | Description |
|---------|------|---------|-------------|
| `formatMode` | string | `"expanded"` | `"expanded"`, `"compact"`, `"preserve"`, or `"detect"` |
| `printWidth` | int | `80` | Max line width for compact/detect modes |
| `experimentalFeatures` | string | `"warning"` | How to handle experimental CSS features: `"ignore"`, `"warning"`, or `"error"` |
| `deprecatedFeatures` | string | `"warning"` | How to handle deprecated (obsolete) CSS features: `"ignore"`, `"warning"`, or `"error"` |

### Experimental Features

Nonstandard CSS properties are filtered out entirely and produce "unknown property" warnings. Experimental properties (e.g. `field-sizing`) are recognized but flagged based on the `experimentalFeatures` setting:

| Value | Diagnostics | Completions |
|-------|------------|-------------|
| `"ignore"` | None | No tagging |
| `"warning"` (default) | Warning severity | Tagged `(experimental)` |
| `"error"` | Error severity | Tagged `(experimental)` |

### Deprecated Features

Deprecated (obsolete) CSS properties (e.g. `clip`) are recognized and flagged based on the `deprecatedFeatures` setting:

| Value | Diagnostics | Completions |
|-------|------------|-------------|
| `"ignore"` | None | No tagging |
| `"warning"` (default) | Warning severity | Tagged `(deprecated)`, strikethrough |
| `"error"` | Error severity | Tagged `(deprecated)`, strikethrough |

## Credits

This project was built using the following as references for feature design and architecture:

- **[microsoft/vscode-css-languageservice](https://github.com/microsoft/vscode-css-languageservice)** (MIT) — CSS language intelligence and LSP feature set. The primary reference for what capabilities a CSS language server should provide.
- **[lmn451/css-variables-zed](https://github.com/lmn451/css-variables-zed)** (GPL-3.0) — workspace-wide CSS variable indexing approach.

## License

MIT — see [LICENSE](LICENSE) for details.
