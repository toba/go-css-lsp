# Go CSS LSP

Language Server Protocol implementation for CSS

## Guidelines

- Be concise
- When fixing or investigating code issues, ALWAYS create a failing test FIRST to demonstrate understanding of the problem THEN change code and confirm the test passes
- Run `golangci-lint run --fix` after modifying Go code
- Run `go test ./...` after changes
- **NEVER commit without explicit user request**

## Building

```bash
go build ./cmd/go-css-lsp
```
