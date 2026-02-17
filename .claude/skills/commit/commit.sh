#!/bin/bash
set -e

# Pre-commit checks
echo "==> Running pre-commit checks..."
golangci-lint run
go test ./...

# Stage and show changes
echo "==> Staging changes..."
git add -A
git status --short
echo ""
echo "==> Staged diff:"
git diff --staged

# Get commit message from arguments
if [ -z "$1" ]; then
    echo ""
    echo "ERROR: Commit subject required as first argument"
    exit 1
fi

SUBJECT="$1"
DESCRIPTION="${2:-}"

# Build commit message
if [ -n "$DESCRIPTION" ]; then
    COMMIT_MSG="$SUBJECT

$DESCRIPTION"
else
    COMMIT_MSG="$SUBJECT"
fi

# Create commit
echo ""
echo "==> Creating commit..."
git commit -m "$COMMIT_MSG"
git status

# Version tagging and release (only if PUSH=true)
if [ "$PUSH" = "true" ]; then
    # Fetch remote tags to ensure we have the latest version info
    echo ""
    echo "==> Fetching remote tags..."
    git fetch --tags

    echo "==> Pushing commits..."
    git push

    # Sync issues to GitHub
    if command -v todo &>/dev/null; then
        echo "==> Syncing issues to GitHub..."
        todo sync
    fi

    CURRENT_TAG=$(git describe --tags --abbrev=0 2>/dev/null || echo "")
    if [ -n "$CURRENT_TAG" ]; then
        echo "==> Current version: $CURRENT_TAG"

        # Auto-bump patch version if NEW_VERSION not explicitly set
        if [ -z "$NEW_VERSION" ]; then
            # Parse current version (e.g., v0.6.0 -> 0 6 0)
            VERSION_NUMS=$(echo "$CURRENT_TAG" | sed 's/^v//' | tr '.' ' ')
            MAJOR=$(echo "$VERSION_NUMS" | awk '{print $1}')
            MINOR=$(echo "$VERSION_NUMS" | awk '{print $2}')
            PATCH=$(echo "$VERSION_NUMS" | awk '{print $3}')
            NEW_VERSION="v${MAJOR}.${MINOR}.$((PATCH + 1))"

            # Check if this version already exists, keep bumping if so
            while git rev-parse "$NEW_VERSION" >/dev/null 2>&1; do
                echo "==> Version $NEW_VERSION already exists, bumping again..."
                PATCH=$((PATCH + 1))
                NEW_VERSION="v${MAJOR}.${MINOR}.$((PATCH + 1))"
            done
            echo "==> Auto-bumping patch version: $CURRENT_TAG -> $NEW_VERSION"
        fi

        echo "==> Creating tag $NEW_VERSION..."
        git tag -a "$NEW_VERSION" -m "Release $NEW_VERSION"

        echo "==> Pushing tag (GoReleaser will create release)..."
        git push origin "$NEW_VERSION"
        echo "==> Tag $NEW_VERSION pushed, GoReleaser workflow will create release"
        echo "==> CI will dispatch version bump to gossamer extension"
    else
        echo "==> No existing tags, skipping version bump"
    fi
else
    echo "==> Commit is local only (use PUSH=true to push and release)"
fi

echo ""
echo "==> Done!"
