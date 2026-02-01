---
name: commit
description: Stage all changes and commit with a descriptive message. Use when the user asks to commit, save changes, or says "/commit".
args: "[push]"
---

## Workflow

**IMPORTANT**: Only use `PUSH=true` when the user explicitly says "/commit push" or asks to push. Plain "/commit" should NEVER push.

1. Review changes to determine commit message and version bump:
   ```bash
   git diff
   git describe --tags --abbrev=0 2>/dev/null || echo "none"
   ```

2. Analyze changes for version bump (if tags exist):
   - **Major (X.0.0)**: Breaking changes - removed/renamed public APIs, changed behavior
   - **Minor (0.X.0)**: New features - new LSP capabilities, new CSS features
   - **Patch (0.0.X)**: Bug fixes, docs, refactoring, dependency updates (auto-bumped if not specified)

3. Run commit script with subject and description:
   ```bash
   # Local commit only (no push, no release):
   .claude/skills/commit/commit.sh "subject line" "description body"

   # Push and release (auto-bumps patch version):
   PUSH=true .claude/skills/commit/commit.sh "subject line" "description body"

   # Push and release with explicit version bump (for major/minor changes):
   PUSH=true NEW_VERSION=vX.Y.Z .claude/skills/commit/commit.sh "subject line" "description body"
   ```

   - **Subject**: Lowercase, imperative mood (e.g., "add feature" not "Added feature")
   - **Description**: Explain the "why" and context. What problem does this solve? What approach was taken? Include relevant details about the implementation.

The script handles: lint (`golangci-lint`), test (`go test`), stage, and commit. Push and release happen when `PUSH=true` (creates tag, GoReleaser builds and publishes binaries).
