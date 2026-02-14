---
# gocssijvk
title: Fix extra space before leading combinator in inline selectors
status: completed
type: bug
priority: normal
created_at: 2026-02-02T00:45:45Z
updated_at: 2026-02-02T00:47:51Z
---

When selectors start with a combinator (e.g. > h1), writeSingleSelector writes a leading space before the combinator. In inline selector lists this produces double spaces like "> header h1,  > h1" instead of "> header h1, > h1".