---
# gocss4bxs
title: Fix DocumentLinkProvider capability type
status: completed
type: bug
priority: normal
created_at: 2026-02-01T22:14:08Z
updated_at: 2026-02-01T22:14:38Z
sync:
    github:
        issue_number: "17"
        synced_at: "2026-02-17T18:03:20Z"
---

The DocumentLinkProvider capability is sent as a boolean `true`, but the Zed client (gossamer) expects a DocumentLinkOptions struct. Change the type from bool to *DocumentLinkOptions.
