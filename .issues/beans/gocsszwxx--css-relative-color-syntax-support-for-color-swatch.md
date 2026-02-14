---
# gocsszwxx
title: CSS relative color syntax support for color swatches
status: completed
type: feature
priority: normal
created_at: 2026-02-01T23:52:11Z
updated_at: 2026-02-02T00:09:38Z
---

Add color swatch support for CSS relative color syntax like rgb(from red r g b), hsl(from #ff0000 h s l), oklch(from green l c h). Requires reverse color-space decomposition, channel name tables, calc() evaluator, origin color parser, and relative color parser.