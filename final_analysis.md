# CSS Formatter Slash Handling Analysis

## Overview
This document describes how the Go CSS LSP formatter handles slashes (/) in property values, particularly for CSS Grid properties like `grid-row: 2/3` and color functions with alpha channels like `rgb(255 128 0 / 50%)`.

## Key Files

### 1. Formatter Main Logic
**File**: `/Users/jason/Developer/toba/go-css-lsp/internal/css/analyzer/format.go`

The formatter is a recursive-descent formatter that handles four formatting modes:
- `FormatExpanded` - One declaration per line (default)
- `FormatCompact` - Single line when fitting PrintWidth
- `FormatPreserve` - Keeps original single/multi-line layout
- `FormatDetect` - Infers intent from source layout

### 2. Scanner/Tokenizer
**File**: `/Users/jason/Developer/toba/go-css-lsp/internal/css/scanner/token.go`
**File**: `/Users/jason/Developer/toba/go-css-lsp/internal/css/scanner/scanner.go`

The scanner tokenizes CSS into a stream of tokens. Importantly:
- Slashes are tokenized as `Delim` tokens (Kind=26)
- The scanner preserves source positions for every token
- Whitespace tokens are created for gaps between tokens

### 3. Parser AST
**File**: `/Users/jason/Developer/toba/go-css-lsp/internal/css/parser/ast.go`

A `Value` node contains:
```go
type Value struct {
    Tokens   []scanner.Token  // All tokens including whitespace
    StartPos int
    EndPos   int
}
```

### 4. Color Analysis
**File**: `/Users/jason/Developer/toba/go-css-lsp/internal/css/analyzer/color.go`

The color analyzer explicitly handles slashes (lines 405-411):
```go
case scanner.Delim:
    switch tok.Value {
    case "/":
        hasSlash = true
    case "-":
        negateNext = true
    }
```

This is used to distinguish modern color syntax (with `/` for alpha) from legacy comma-separated syntax.

## Current Slash Handling Behavior

### The Key Function: `writeValueTo` (lines 749-773)

```go
func (f *formatter) writeValueTo(
    sb *strings.Builder,
    v *parser.Value,
) {
    prevEnd := -1
    prevKind := scanner.EOF
    for _, tok := range v.Tokens {
        if tok.Kind == scanner.Whitespace {
            continue  // Skip whitespace tokens
        }
        // Add space if there's a gap in source, UNLESS:
        if prevEnd >= 0 && tok.Offset > prevEnd {
            if tok.Kind != scanner.ParenClose &&
                prevKind != scanner.Function &&
                prevKind != scanner.ParenOpen {
                sb.WriteByte(' ')
            }
        }
        sb.WriteString(
            string(f.src[tok.Offset:tok.End]),
        )
        prevEnd = tok.End
        prevKind = tok.Kind
    }
}
```

### How It Works

The formatter:
1. **Skips all Whitespace tokens** (source whitespace is not preserved)
2. **Detects gaps in the source** by comparing `prevEnd` with `tok.Offset`
3. **Adds a space** when there's a gap, UNLESS the token is `)` or follows `(` or a function name
4. **Preserves exact tokens** from source (via slicing `f.src[tok.Offset:tok.End]`)

### Behavior with Slashes

**Example 1: `grid-row: 2/3`**
- Token sequence: `Number("2")` → `Delim("/")` → `Number("3")`
- Token offsets: `[18,19]` → `[19,20]` → `[20,21]`
- No gaps between tokens (prevEnd == tok.Offset for all)
- **Result**: `2/3` (no spaces added)

**Example 2: `grid-row: 2 / 3`**
- Token sequence: `Number("2")` → `Whitespace` → `Delim("/")` → `Whitespace` → `Number("3")`
- Whitespace tokens are skipped, so effective sequence: `Number("2")` → `Delim("/")` → `Number("3")`
- Token offsets: `[18,19]` → `[38,39]` → `[40,41]`
- Gaps exist (37 to 38, 39 to 40)
- **Result**: `1 / 4` (spaces added to preserve original intent)

**Example 3: `rgb(255 128 0 / 50%)`**
- Tokens: `Function("rgb")` → `Number` → `Number` → `Number` → `Delim("/")` → `Percentage`
- Whitespace skipped
- Gaps exist around `/`
- **Result**: `rgb(255 128 0 / 50%)` (spaces preserved, but no space after opening paren due to special case)

## Slash Tokenization Details

The slash is always tokenized as a `Delim` token because:
1. The CSS Scanner spec treats `/` as a delimiter character
2. The scanner's tokenization logic (lines 44-46) specifically checks for comments starting with `/*`, but a single `/` is just a delimiter
3. Slashes are NOT given special token types like `Comma` or `Colon`

## Special Cases

### Slashes NOT Treated as Operators
- No arithmetic (division) support - `/` is literally preserved
- No special spacing rules for `/` - it's treated like any other delimiter

### Spacing Rules That Apply to Slashes
1. **Between ParenOpen and next token**: No space added
2. **Between previous token and ParenClose**: No space added  
3. **Between other tokens with gaps**: Space added if source had gap

### No Special Slash Handling in Formatter
Currently, the formatter has:
- NO special case for `Delim` token kind
- NO special case for `/` value
- NO explicit grid syntax support
- NO explicit color syntax support

The formatter simply preserves whatever spacing was in the source input.

## Test Coverage

Current tests in `/Users/jason/Developer/toba/go-css-lsp/internal/css/analyzer/format_test.go`:
- No tests for grid properties with slashes
- No tests for color functions with alpha slashes
- Many tests for function parens, commas, and breakpoints
- No tests for `Delim` token spacing

## Tokenization Example Output

For `rgb(255,128,0/50%)`:
```
Token 0: Function("rgb") Kind=13
Token 1: Number("255")  Kind=8
Token 2: Comma          Kind=19  (offset=7, end=8)
Token 3: Number("128")  Kind=8   (offset=8, end=11)
Token 4: Comma          Kind=19  (offset=11, end=12)
Token 5: Number("0")    Kind=8   (offset=12, end=13)
Token 6: Delim("/")     Kind=26  (offset=13, end=14)
Token 7: Percentage     Kind=9   (offset=14, end=17)
Token 8: ParenClose     Kind=23  (offset=17, end=18)
```

Notice: Delim "/" is token kind 26, same as other single-character delimiters.

## Architecture Notes

- **Token preservation**: The formatter reconstructs output by slicing original source bytes, not by building tokens from scratch
- **Gap detection**: Formatting decisions are based on byte offset gaps in the original source
- **Delimiter agnostic**: The formatter doesn't distinguish between different delimiters (/, -, +, etc.)
- **No token reconstruction**: Delimiters are never recreated - they're always from the source via `f.src[tok.Offset:tok.End]`
