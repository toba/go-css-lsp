---
# gocssa36r
title: Fix extra space after opening paren in value formatting
status: completed
type: bug
priority: normal
created_at: 2026-02-01T22:59:19Z
updated_at: 2026-02-01T23:01:47Z
---

writeValue/writeValueTo insert a space after Function and ParenOpen tokens because they only check for ParenClose suppression. Need to also suppress space when the previous token was Function or ParenOpen.