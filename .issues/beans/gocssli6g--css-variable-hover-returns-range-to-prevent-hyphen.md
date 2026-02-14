---
# gocssli6g
title: CSS variable hover returns range to prevent hyphen-split highlights
status: completed
type: bug
priority: normal
created_at: 2026-02-01T23:50:54Z
updated_at: 2026-02-01T23:53:58Z
---

When cmd+hovering a CSS variable like var(--color-link-icon), Zed highlights each hyphen-separated segment individually. The hover response needs a Range field so the editor uses that instead of its own word detection.