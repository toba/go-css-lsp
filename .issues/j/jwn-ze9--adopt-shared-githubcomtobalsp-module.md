---
# jwn-ze9
title: Adopt shared github.com/toba/lsp module
status: completed
type: feature
priority: normal
created_at: 2026-03-21T17:18:32Z
updated_at: 2026-03-21T17:33:07Z
sync:
    github:
        issue_number: "71"
        synced_at: "2026-03-21T17:45:55Z"
---

Replace duplicated LSP infrastructure code with the shared github.com/toba/lsp module. This eliminates ~500-700 lines of copy-pasted code that is identical across all four toba LSP projects.

## Packages to adopt

- `github.com/toba/lsp/transport` — replaces hand-rolled `parsing.go` (ReceiveInput, decode, Encode, SendToLspClient, SendOutput)
- `github.com/toba/lsp/logging` — replaces createLogFile(), configureLogging(), MaxLogFileSize/DirPermissions/FilePermissions constants
- `github.com/toba/lsp/pathutil` — replaces uriToFilePath(), filePathToUri(), convertKeysFromFilePathToUri()
- `github.com/toba/lsp/position` — replaces OffsetToLineChar/offsetToLineCol, LineCharToOffset, intToUint, uintToInt

## Steps

1. ~~Add `github.com/toba/lsp` dependency~~ done
2. ~~Replace transport layer~~ done — deleted `parsing.go`, using `transport.NewScanner` and `transport.Send`
3. ~~Replace logging~~ done — using `logging.Configure(serverName)`
4. ~~Replace path utilities~~ done — using `pathutil.URIToFilePath`
5. ~~Replace position utilities~~ done — `analyzer.position.go` delegates to `position.OffsetToLineCol`/`LineColToOffset`
6. ~~Delete replaced local code~~ done
7. ~~Remove unused constants~~ done
8. ~~Run tests and linter~~ done — all pass, 0 lint issues


## Summary of Changes

Adopted `github.com/toba/lsp` v0.1.1 shared module:
- **transport**: Deleted `parsing.go` (+ tests), replaced with `transport.NewScanner`/`transport.Send`
- **logging**: Deleted `createLogFile`/`configureLogging`, replaced with `logging.Configure(serverName)`
- **pathutil**: Deleted local `uriToFilePath`, replaced with `pathutil.URIToFilePath`
- **position**: `analyzer/position.go` now delegates to `position.OffsetToLineCol`/`LineColToOffset`
- Removed unused constants (`ContentLengthHeader`, `HeaderDelimiter`, `LineDelimiter`, `DirPermissions`, `FilePermissions`, `MaxLogFileSize`)
- Renamed module from `github.com/toba/go-css-lsp` to `github.com/toba/css-lsp`
- All tests pass, zero lint issues
