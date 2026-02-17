---
# gocss3nd4
title: Fix concurrent stdout writes corrupting LSP protocol stream
status: completed
type: bug
priority: normal
created_at: 2026-02-02T16:41:10Z
updated_at: 2026-02-02T16:43:34Z
sync:
    github:
        issue_number: "14"
        synced_at: "2026-02-17T18:03:20Z"
---

Main goroutine and diagnostic goroutine both call lsp.SendToLspClient(os.Stdout, ...) without synchronization, causing interleaved writes that corrupt Content-Length framing. Add a sync.Mutex to protect all stdout writes.
