---
# gocss9gbj
title: Enhanced lint rules
status: completed
type: feature
priority: normal
created_at: 2026-02-01T20:56:17Z
updated_at: 2026-02-01T21:13:00Z
parent: gocsse5ea
sync:
    github:
        issue_number: "61"
        synced_at: "2026-02-17T18:03:22Z"
---

Add additional diagnostic/lint rules matching vscode-css-languageservice.

## Reference
- vscode-css-languageservice: lint.ts, lintRules.ts

## Checklist
- [ ] Vendor prefix warnings
- [ ] Zero-unit detection (0px → 0)
- [ ] Hex color length validation
- [ ] Box model size warnings (width + padding)
- [ ] Universal selector performance warnings
- [ ] !important usage warnings
- [ ] Float usage warnings
- [ ] @import performance warnings
- [ ] Properties ignored due to display value
- [ ] Configurable lint severity (error/warning/info/ignore)
- [ ] Add tests
