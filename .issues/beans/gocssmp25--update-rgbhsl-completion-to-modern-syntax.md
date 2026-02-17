---
# gocssmp25
title: Update rgb()/hsl() completion to modern syntax
status: completed
type: bug
priority: normal
created_at: 2026-02-01T21:47:39Z
updated_at: 2026-02-01T21:51:48Z
sync:
    github:
        issue_number: "4"
        synced_at: "2026-02-17T18:03:20Z"
---

Port of vscode-css-languageservice #413. Color function completions use old comma-separated syntax (e.g. rgb(r, g, b)) instead of the modern space-separated syntax (e.g. rgb(r g b / a)). Update completion snippets to use the current CSS spec syntax.
