# Changelog

## Week of Mar 1 – Mar 7, 2026

### 🗜️ Tweaks

- Update vscode-custom-data submodule to v0.6.3; 37 new properties, updated baseline/experimental status ([#68](https://github.com/toba/go-css-lsp/issues/68))

## Week of Feb 22 – Feb 28, 2026

### ✨ Features

- Support `@scope` selector lists ([#65](https://github.com/toba/go-css-lsp/issues/65))
- Support CSS `if()` function ([#67](https://github.com/toba/go-css-lsp/issues/67))

### 🐞 Fixes

- Fix `@container` query parsing ([#66](https://github.com/toba/go-css-lsp/issues/66))

## Week of Feb 15 – Feb 21, 2026

### ✨ Features

- Incorporate upstream vscode-css-languageservice features ([#27](https://github.com/toba/go-css-lsp/issues/27))

## Week of Feb 1 – Feb 7, 2026

### ✨ Features

- Color swatches for CSS variables ([#30](https://github.com/toba/go-css-lsp/issues/30))
- Document symbols ([#55](https://github.com/toba/go-css-lsp/issues/55))
- Code actions / quick fixes ([#62](https://github.com/toba/go-css-lsp/issues/62))
- Document colors and color presentations ([#58](https://github.com/toba/go-css-lsp/issues/58))
- Find references ([#56](https://github.com/toba/go-css-lsp/issues/56))
- Improve function hover; signatures + MDN links ([#7](https://github.com/toba/go-css-lsp/issues/7))
- Implement compact formatting modes for CSS formatter ([#5](https://github.com/toba/go-css-lsp/issues/5))
- CSS value validation and completion fix ([#42](https://github.com/toba/go-css-lsp/issues/42))
- Inline value prefix when wrapping inside functions ([#46](https://github.com/toba/go-css-lsp/issues/46))
- CSS formatting ([#59](https://github.com/toba/go-css-lsp/issues/59))
- Add CSS status/deprecation awareness from BCD ([#25](https://github.com/toba/go-css-lsp/issues/25))
- Break long property values at top-level commas ([#33](https://github.com/toba/go-css-lsp/issues/33))
- Enhanced lint rules ([#61](https://github.com/toba/go-css-lsp/issues/61))
- Color conversion code actions ([#11](https://github.com/toba/go-css-lsp/issues/11))
- Add `FormatDetect` mode + blank line handling for compact/preserve/detect ([#43](https://github.com/toba/go-css-lsp/issues/43))
- Document highlights ([#57](https://github.com/toba/go-css-lsp/issues/57))
- CSS relative color syntax support for color swatches ([#36](https://github.com/toba/go-css-lsp/issues/36))
- Rename support ([#54](https://github.com/toba/go-css-lsp/issues/54))
- Add experimental features config setting ([#45](https://github.com/toba/go-css-lsp/issues/45))
- Workspace-wide CSS variable indexing ([#63](https://github.com/toba/go-css-lsp/issues/63))
- Go to definition for CSS variables ([#51](https://github.com/toba/go-css-lsp/issues/51))
- Add `source.fixAll` code actions for auto-fix on save ([#31](https://github.com/toba/go-css-lsp/issues/31))
- Add `deprecatedFeatures` setting ([#37](https://github.com/toba/go-css-lsp/issues/37))
- Selection ranges ([#52](https://github.com/toba/go-css-lsp/issues/52))
- Return `LocationLink[]` from `textDocument/definition` for proper cmd-click underline ([#41](https://github.com/toba/go-css-lsp/issues/41))
- Document links ([#53](https://github.com/toba/go-css-lsp/issues/53))
- Detect selector list formatting in detect mode ([#8](https://github.com/toba/go-css-lsp/issues/8))
- Folding ranges ([#60](https://github.com/toba/go-css-lsp/issues/60))

### 🐞 Fixes

- Fix `DocumentLinkProvider` capability type ([#17](https://github.com/toba/go-css-lsp/issues/17))
- Fix completions triggering inside comments ([#49](https://github.com/toba/go-css-lsp/issues/49))
- Exclude `currentcolor` from color swatches ([#35](https://github.com/toba/go-css-lsp/issues/35))
- Cross-file variable hover shows value ([#3](https://github.com/toba/go-css-lsp/issues/3))
- Support single-line at-rules in detect/compact/preserve format modes ([#32](https://github.com/toba/go-css-lsp/issues/32))
- Long lines not wrapped in detect formatting mode ([#20](https://github.com/toba/go-css-lsp/issues/20))
- Formatter removes single blank line between declarations with detect option ([#50](https://github.com/toba/go-css-lsp/issues/50))
- Fix remaining unknown-value false positives ([#38](https://github.com/toba/go-css-lsp/issues/38))
- Parse modern color functions for color picker ([#24](https://github.com/toba/go-css-lsp/issues/24))
- CSS variable hover returns range to prevent hyphen-split highlights ([#18](https://github.com/toba/go-css-lsp/issues/18))
- Fix extra space after opening paren in value formatting ([#15](https://github.com/toba/go-css-lsp/issues/15))
- Fix false positives; transition/animation property names flagged as unknown values ([#28](https://github.com/toba/go-css-lsp/issues/28))
- Support CSS nesting; `&` selector, nested rules ([#19](https://github.com/toba/go-css-lsp/issues/19))
- Fix concurrent stdout writes corrupting LSP protocol stream ([#14](https://github.com/toba/go-css-lsp/issues/14))
- Fix LSP server crashes from unrecovered panics ([#39](https://github.com/toba/go-css-lsp/issues/39))
- Detect format mode not wired up in LSP server ([#16](https://github.com/toba/go-css-lsp/issues/16))
- Fix concurrent map access crash in LSP server ([#26](https://github.com/toba/go-css-lsp/issues/26))
- Fix CSS variable cmd-click highlight range and cross-file navigation ([#1](https://github.com/toba/go-css-lsp/issues/1))
- Fix color swatches for chained CSS variable references ([#34](https://github.com/toba/go-css-lsp/issues/34))
- Fix extra space before leading combinator in inline selectors ([#12](https://github.com/toba/go-css-lsp/issues/12))
- Update `rgb()`/`hsl()` completion to modern syntax ([#4](https://github.com/toba/go-css-lsp/issues/4))
- Preserve blank lines before comments in ruleset bodies ([#48](https://github.com/toba/go-css-lsp/issues/48))
- Go to Definition for CSS variables across files ([#2](https://github.com/toba/go-css-lsp/issues/2))
- Fix false unknown-value diagnostics for `background` shorthand ([#64](https://github.com/toba/go-css-lsp/issues/64))

### 🗜️ Tweaks

- CSS3 Language Server; feature parity ([#44](https://github.com/toba/go-css-lsp/issues/44))
- Go optimization report for full codebase ([#6](https://github.com/toba/go-css-lsp/issues/6))
