---
# gocssgu7a
title: Support single-line at-rules in detect/compact/preserve format modes
status: completed
type: bug
priority: normal
created_at: 2026-02-02T00:35:27Z
updated_at: 2026-02-02T00:38:50Z
sync:
    github:
        issue_number: "32"
        synced_at: "2026-02-17T18:03:20Z"
---

formatAtRule always formats multi-line. When an @media rule (or similar) only contains declarations and fits within PrintWidth, detect/compact/preserve modes should be able to collapse it to a single line, just like rulesets.
