---
# qx1-2le
title: Review upstream custom data update from vscode-css-languageservice
status: completed
type: task
priority: normal
created_at: 2026-03-03T17:54:46Z
updated_at: 2026-03-03T18:00:26Z
sync:
    github:
        issue_number: "68"
        synced_at: "2026-03-03T18:02:04Z"
---

Upstream commit 3cc5015 ("update custom data #476") updated `src/data/webCustomData.ts` and `src/languageFacts/builtinData.ts`. Check if our generated CSS data from vscode-custom-data needs a refresh.

## Todo

- [x] Compare upstream `webCustomData.ts` changes against our generated data
- [x] Re-run data generation if needed (`jig todo show zjj-3g8` for prior process reference)
- [x] Verify no new properties/values are missing from completions or validation


## Summary of Changes

Updated vscode-custom-data submodule from v0.6.2 to v0.6.3 and regenerated CSS data. Properties grew from 564 to 601 (+37 new properties including corner-*-shape, animation-trigger, timeline-trigger-*, reading-flow, scroll-marker-group). Updated hover test to use initial-letter-align instead of field-sizing which is no longer experimental upstream. All tests pass.
