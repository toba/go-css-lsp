---
# gocss5ube
title: Exclude currentcolor from color swatches
status: completed
type: bug
priority: normal
created_at: 2026-02-02T00:30:39Z
updated_at: 2026-02-02T00:32:44Z
---

currentcolor is included in namedColorMap and gets a swatch with a placeholder black color. It should be excluded since it's not a concrete color value.