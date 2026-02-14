---
# gocssg5fm
title: Formatter removes single blank line between declarations with detect option
status: completed
type: bug
priority: normal
created_at: 2026-02-02T16:58:00Z
updated_at: 2026-02-02T17:00:05Z
---

When using the detect blank line option, the formatter removes single blank lines between CSS declarations. Expected: a single blank line between declarations should be preserved. Example:\n\n```css\n--button-icon-size: calc(var(--button-size) - var(--button-radius) * 2);\n\nposition: relative;\n```\n\nThe formatter strips the blank line.