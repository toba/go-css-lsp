---
# naw-j90
title: Incorporate upstream vscode-css-languageservice features
status: completed
type: feature
priority: normal
created_at: 2026-02-16T21:53:54Z
updated_at: 2026-02-16T21:58:25Z
---

Two features from upstream:

1. Media query feature auto-completion — complete media features like width, prefers-color-scheme inside @media ()
2. Baseline status in hover — show browser availability info on hover

## Tasks
- [x] Update generate-data to parse baseline + media descriptors
- [x] Add Baseline and MediaFeature types to data package
- [x] Regenerate data files
- [x] Add media feature/value completion to analyzer
- [x] Add baseline display to hover handlers
- [x] Add completion tests
- [x] Add hover tests
- [x] Run tests and lint

## Summary of Changes

### Feature 1: Media Query Feature Completion
- Added `cssDescriptor` and `cssBaseline` structs to code generator
- Generator now extracts `@media` descriptors into `media_features_gen.go` (38 media features)
- Added `MediaFeature` type and lookup functions to data package
- Added `contextMediaFeature` and `contextMediaValue` completion contexts
- `detectMediaContext()` scans for `@media (...)` pattern and determines feature vs value context
- `completeMediaFeatures()` and `completeMediaValues()` provide completions

### Feature 2: Baseline Status in Hover
- Added `Baseline` struct to data package with Status, LowDate, HighDate
- Generator now emits baseline info for properties, at-rules, pseudo-classes, and pseudo-elements
- `baselineLabel()` helper renders human-readable availability labels
- Baseline info shown in hover for properties, at-rules, pseudo-classes, and pseudo-elements
- Baseline suppressed for experimental features (matching upstream behavior)
