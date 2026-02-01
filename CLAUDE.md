# Go CSS LSP

Language Server Protocol implementation for CSS3 (CSS3 only — no SCSS, Sass, or LESS support).

## Reference Projects

- [microsoft/vscode-css-languageservice](https://github.com/microsoft/vscode-css-languageservice) — feature reference for LSP capabilities
- [lmn451/css-variables-zed](https://github.com/lmn451/css-variables-zed) — reference for CSS variable workspace indexing
- Companion Zed extension: `../gossamer`

## Guidelines

- Be concise
- CSS3 support only — no SCSS, Sass, or LESS
- When fixing or investigating code issues, ALWAYS create a failing test FIRST to demonstrate understanding of the problem THEN change code and confirm the test passes
- Run `golangci-lint run --fix` after modifying Go code
- Run `go test ./...` after changes
- **NEVER commit without explicit user request**

## Building

```bash
go build ./cmd/go-css-lsp
```
