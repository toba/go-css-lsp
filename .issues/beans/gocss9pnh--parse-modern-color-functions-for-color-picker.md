---
# gocss9pnh
title: Parse modern color functions for color picker
status: completed
type: bug
priority: normal
created_at: 2026-02-01T21:47:39Z
updated_at: 2026-02-01T21:51:48Z
sync:
    github:
        issue_number: "24"
        synced_at: "2026-02-17T18:03:20Z"
---

Port of vscode-css-languageservice #305/#314/#456. The color picker only works for rgb(), rgba(), hsl(), hsla(), hwb(), hex, and named colors. Modern CSS color functions lab(), lch(), oklch(), oklab() are listed in ColorFunctions but have no parsing logic in color.go. Also space-separated rgb/hsl values (e.g. 'rgb(255 255 255 / 80%)') need color picker support. Add parsing for all modern color spaces.
