---
# gocss0zn6
title: Fix completions triggering inside comments
status: completed
type: bug
priority: normal
created_at: 2026-02-01T21:47:39Z
updated_at: 2026-02-01T21:51:48Z
---

Port of vscode-css-languageservice #429. Pseudo-class/element and at-rule suggestions are triggered when typing inside CSS comments. The determineContext() function in completion.go does naive backward string searches without checking if the offset falls within a Comment node boundary. Fix: check if offset is inside a comment before returning pseudo-class/at-rule context.