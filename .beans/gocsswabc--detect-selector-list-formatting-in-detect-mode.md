---
# gocsswabc
title: Detect selector list formatting in detect mode
status: completed
type: feature
priority: normal
created_at: 2026-02-01T23:44:26Z
updated_at: 2026-02-01T23:46:11Z
---

In detect formatting mode, selector lists should behave like rules: if the second selector is on the same line as the first, all selectors should be inline (if they fit within print width). If the second selector is on a new line, all should be on new lines.