#!/usr/bin/env bash
# Run golangci-lint with auto-fixes, then report remaining issues.
set -euo pipefail

repo_root="$(git -C "$(dirname "$0")" rev-parse --show-toplevel)" || exit 1
cd "${repo_root}"

echo "Running golangci-lint with auto-fix..."
golangci-lint run --fix ./... 2>&1 || true

echo ""
echo "Remaining issues (require manual intervention):"
golangci-lint run ./... 2>&1 || true
