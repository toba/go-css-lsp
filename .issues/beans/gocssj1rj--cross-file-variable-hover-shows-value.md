---
# gocssj1rj
title: Cross-file variable hover shows value
status: completed
type: bug
priority: normal
created_at: 2026-02-02T00:36:17Z
updated_at: 2026-02-02T00:38:05Z
sync:
    github:
        issue_number: "3"
        synced_at: "2026-02-17T18:03:20Z"
---

Hovering CSS variables defined in other files doesn't show their value in the popover, while same-file variables do. The hover path needs to accept a VariableResolver to fall back to the workspace index when the variable isn't found in the current file's AST.
