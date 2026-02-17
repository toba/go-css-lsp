---
# gocss6qx4
title: Add tests for concurrent stdout write protection
status: completed
type: task
priority: normal
created_at: 2026-02-02T16:48:06Z
updated_at: 2026-02-02T16:50:07Z
sync:
    github:
        issue_number: "29"
        synced_at: "2026-02-17T18:03:20Z"
---

Add two tests: (1) a race detector test that spawns concurrent writers to verify no data race on the shared writer, and (2) a protocol framing validation test that verifies concurrent writes produce valid Content-Length framed messages.
