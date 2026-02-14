---
# gocss2tm3
title: Break long property values at top-level commas
status: completed
type: feature
priority: normal
created_at: 2026-02-01T22:56:49Z
updated_at: 2026-02-01T22:57:49Z
---

When a declaration line exceeds print width and the value contains top-level commas (not inside parentheses), break the value across multiple lines at those commas. Property name goes on its own line with colon, then each comma-separated segment on a continuation line indented one level deeper.