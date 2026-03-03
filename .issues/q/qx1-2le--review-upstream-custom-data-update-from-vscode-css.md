---
# qx1-2le
title: Review upstream custom data update from vscode-css-languageservice
status: ready
type: task
priority: normal
created_at: 2026-03-03T17:54:46Z
updated_at: 2026-03-03T17:54:46Z
sync:
    github:
        issue_number: "68"
        synced_at: "2026-03-03T17:56:40Z"
---

Upstream commit 3cc5015 ("update custom data #476") updated `src/data/webCustomData.ts` and `src/languageFacts/builtinData.ts`. Check if our generated CSS data from vscode-custom-data needs a refresh.

## Todo

- [ ] Compare upstream `webCustomData.ts` changes against our generated data
- [ ] Re-run data generation if needed (`jig todo show zjj-3g8` for prior process reference)
- [ ] Verify no new properties/values are missing from completions or validation
