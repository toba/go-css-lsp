---
# zjj-nt8
title: Formatter appends stray '} {}' to end of file
status: ready
type: bug
priority: high
created_at: 2026-02-19T19:41:35Z
updated_at: 2026-02-19T19:59:27Z
sync:
    github:
        issue_number: "64"
        synced_at: "2026-02-21T05:50:00Z"
---

Auto format is routinely appending `} {}` after the final closing brace of CSS files. This has caused multiple deploy failures.

Example from `web/portfolio/portfolio-list.css`:
```
108        }
109     }
110  }
111 -} {}
```

The formatter is emitting an extra `} {}` token sequence at the end of the output.

## Investigation Findings

### Mechanism
`} {}` is a phantom empty ruleset where `}` is the "selector". `buildSingleLine` (format.go:714-747) renders a Ruleset with no children and a `}` selector token as `} {}`. In expanded mode it would be `} {\n}\n`.

### Root cause (partial)
`parseSelector` (parser.go:382-445) only breaks on `EOF`, `BraceOpen`, and `Comma`. It does **not** break on `BraceClose`. So if a stray `}` reaches `parseStylesheet`, it gets consumed as a selector part, and the next token boundary creates a phantom Ruleset node.

### What's NOT the cause
- **LSP edit range**: `processFormatting` (main.go:726-778) replaces `0..len(src)` — full document, no off-by-one.
- **`FormatDocument`**: thin wrapper over `analyzer.Format`, no extra logic.
- **`parseBlock` depth tracking**: depth starts at 1, child parsers consume their own braces, so it always goes 1→0 correctly.
- **`parseRuleset` brace consumption**: correctly consumes `}` in its block loop (parser.go:306-309).
- **`parseAtRule` → `parseBlock`**: correctly dispatches to child parsers and consumes the block's `}`.

### What's still unknown
The exact CSS input pattern that causes a `}` to leak to the top level. Extensive tracing of nested patterns (media queries, CSS nesting with `&`, `@layer`, empty rulesets, 4-level nesting, compact/no-newline variants) all parse correctly. Tested 11 CSS patterns × 4 format modes — none reproduced `} {}`.

### Possible next steps
1. Add logging/assertions to the parser to detect when `parseStylesheet` encounters `BraceClose`
2. Get the actual `portfolio-list.css` source that triggers it
3. Test idempotency (format output → reparse → format again)
4. Check if the Zed extension preprocesses the source before sending to LSP
5. Regardless of trigger: harden `parseStylesheet` to skip stray `}` tokens instead of passing them to `parseRuleset`

### Test file
Exploratory test added at `internal/css/analyzer/format_stray_test.go` — currently all passing (no reproduction yet).

## Tasks

- [ ] Reproduce the stray `} {}` with a failing test
- [ ] Fix parser or formatter
- [ ] Confirm test passes and no regressions
