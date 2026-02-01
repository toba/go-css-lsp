---
# gocssofi8
title: Add FormatDetect mode + blank line handling for compact/preserve/detect
status: completed
type: feature
priority: normal
created_at: 2026-02-01T23:02:54Z
updated_at: 2026-02-01T23:05:08Z
---

## Overview

Two changes to the CSS formatter:
1. **Detect mode**: new FormatDetect mode that checks if first property is inline with `{` — if so and fits print-width, single-line; otherwise multi-line
2. **Compact, preserve, detect**: don't force blank lines between rules; preserve/detect respects original spacing (capped at 1 blank line); compact always 0

## Checklist

- [ ] Add FormatDetect constant
- [ ] Add isFirstChildInline helper
- [ ] Add formatRulesetDetect method
- [ ] Add writePreservedBlankLines helper
- [ ] Update formatStylesheet for mode-conditional blank lines
- [ ] Update formatRuleset for nested rule blank lines
- [ ] Update formatAtRule for at-rule block blank lines
- [ ] Add preserve mode print-width gate to formatRulesetPreserve
- [ ] Update existing tests (MixedRules, NestedFallsBack variants)
- [ ] Add new tests for detect mode
- [ ] Add new tests for blank line handling
- [ ] Run go test ./... and golangci-lint run --fix