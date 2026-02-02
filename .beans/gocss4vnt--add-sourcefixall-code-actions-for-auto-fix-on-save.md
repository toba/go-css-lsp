---
# gocss4vnt
title: Add source.fixAll code actions for auto-fix on save
status: completed
type: feature
priority: normal
created_at: 2026-02-02T00:35:29Z
updated_at: 2026-02-02T00:38:03Z
---

Add source.fixAll support so LSP clients can auto-fix simple diagnostics on save. Currently the only auto-fixable diagnostic is unnecessary unit (0deg→0, 0px→0, etc.).