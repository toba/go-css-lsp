---
# gocssubam
title: Preserve blank lines before comments in ruleset bodies
status: completed
type: bug
priority: normal
created_at: 2026-02-02T01:18:58Z
updated_at: 2026-02-02T01:20:12Z
---

The formatter strips blank lines before comments inside ruleset bodies. Need to add blank line preservation for Comment nodes in formatRulesetBody().