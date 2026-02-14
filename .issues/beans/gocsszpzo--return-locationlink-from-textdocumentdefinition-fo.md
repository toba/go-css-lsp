---
# gocsszpzo
title: Return LocationLink[] from textDocument/definition for proper cmd-click underline
status: completed
type: feature
priority: normal
created_at: 2026-02-02T00:32:52Z
updated_at: 2026-02-02T00:34:14Z
---

When cmd-clicking a CSS variable like var(--color-link-icon), Zed highlights each hyphen-separated segment individually instead of the full variable name. Return LocationLink[] with originSelectionRange instead of Location to fix this.

## Checklist
- [ ] Add LocationLink type to lsp/methods.go
- [ ] Add DefinitionResult struct to analyzer/definition.go with origin range
- [ ] Update FindDefinition to return origin range (var() span)
- [ ] Update css.Definition to return DefinitionResult
- [ ] Update processDefinition in main.go to build LocationLink response
- [ ] Update definition_test.go to verify origin range
- [ ] Run tests and linter