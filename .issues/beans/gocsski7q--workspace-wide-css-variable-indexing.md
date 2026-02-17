---
# gocsski7q
title: Workspace-wide CSS variable indexing
status: completed
type: feature
priority: normal
created_at: 2026-02-01T20:56:17Z
updated_at: 2026-02-01T21:19:10Z
parent: gocsse5ea
sync:
    github:
        issue_number: "63"
        synced_at: "2026-02-17T18:03:22Z"
---

Index CSS custom properties across all workspace CSS files for cross-file go-to-definition, references, and completion.

## Reference
- lmn451/css-variables-zed: workspace indexing with file watcher

## Checklist
- [ ] Scan workspace for .css files on initialization
- [ ] Build index of all custom property definitions with file/position
- [ ] Watch for file changes (workspace/didChangeWatchedFiles)
- [ ] Cross-file go-to-definition for var(--name)
- [ ] Cross-file completion for var(--name)
- [ ] Cross-file find-references
- [ ] Add tests
