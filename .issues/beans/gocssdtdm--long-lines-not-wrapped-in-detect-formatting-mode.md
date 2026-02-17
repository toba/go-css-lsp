---
# gocssdtdm
title: Long lines not wrapped in detect formatting mode
status: completed
type: bug
priority: normal
created_at: 2026-02-01T23:40:18Z
updated_at: 2026-02-01T23:44:53Z
sync:
    github:
        issue_number: "20"
        synced_at: "2026-02-17T18:03:20Z"
---

In formatMode detect, long lines should still be wrapped if they exceed printWidth. Example line that stays unwrapped:

```css
--color-bg-top: light-dark(hsl(from var(--brand-deep-pockets) h calc(s - 40) calc(l + 40)), hsl(from var(--brand-deep-pockets) h calc(s - 40) l));
```

## Checklist
- [ ] Create a failing test demonstrating the issue
- [ ] Fix the formatting logic to wrap long lines even in detect mode
- [ ] Verify tests pass
