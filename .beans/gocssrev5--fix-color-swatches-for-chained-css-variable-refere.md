---
# gocssrev5
title: Fix color swatches for chained CSS variable references
status: completed
type: bug
priority: normal
created_at: 2026-02-03T00:39:48Z
updated_at: 2026-02-03T00:40:29Z
---

var(--alias) where --alias: var(--blue) and --blue: #0000ff doesn't produce a color swatch. The resolveVarColor function passes nil as the resolver when recursing, preventing any chained resolution.