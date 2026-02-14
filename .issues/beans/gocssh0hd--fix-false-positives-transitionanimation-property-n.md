---
# gocssh0hd
title: 'Fix false positives: transition/animation property names flagged as unknown values'
status: completed
type: bug
priority: normal
created_at: 2026-02-02T01:05:00Z
updated_at: 2026-02-02T01:05:55Z
---

transition: border-color 0.15s produces false unknown value diagnostics. Fix checkUnknownValues to accept CSS property names for transition/transition-property/will-change, and skip validation for animation-name/animation.