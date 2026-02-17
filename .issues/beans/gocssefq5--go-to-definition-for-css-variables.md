---
# gocssefq5
title: Go to definition for CSS variables
status: completed
type: feature
priority: high
created_at: 2026-02-01T20:56:16Z
updated_at: 2026-02-01T21:06:48Z
parent: gocsse5ea
blocking:
    - gocssnrkb
    - gocsstxry
    - gocssd88w
sync:
    github:
        issue_number: "51"
        synced_at: "2026-02-17T18:03:22Z"
---

Navigate from var(--name) usage to the --name definition within the same document (single-file first, workspace later).

## Reference
- vscode-css-languageservice: cssNavigation.ts (findDefinition)
- lmn451/css-variables-zed: workspace-wide variable indexing

## Checklist
- [ ] Find custom property definitions in current document
- [ ] Resolve var(--name) to its declaration
- [ ] Register textDocument/definition capability
- [ ] Add tests
- [ ] (Future) Workspace-wide variable index for cross-file definitions
