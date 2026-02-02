package analyzer

import (
	"strconv"

	"github.com/toba/go-css-lsp/internal/css/scanner"
)

// isMinusToken returns true if the token represents a minus
// sign. The CSS scanner may tokenize "-" as either Delim or
// Ident depending on context.
func isMinusToken(tok scanner.Token) bool {
	return (tok.Kind == scanner.Delim ||
		tok.Kind == scanner.Ident) && tok.Value == "-"
}

// isAddSubOp checks if a token is an additive operator.
func isAddSubOp(tok scanner.Token) (string, bool) {
	if tok.Kind == scanner.Delim && tok.Value == "+" {
		return "+", true
	}
	if isMinusToken(tok) {
		return "-", true
	}
	return "", false
}

// evalCalc evaluates a simple calc()-style expression from CSS
// tokens. channelVars maps channel name identifiers (like "r",
// "s", "l") to their numeric values. Returns the result and
// whether evaluation succeeded.
func evalCalc(
	tokens []scanner.Token,
	channelVars map[string]float64,
) (float64, bool) {
	// Filter out whitespace tokens.
	var filtered []scanner.Token
	for _, t := range tokens {
		if t.Kind != scanner.Whitespace {
			filtered = append(filtered, t)
		}
	}
	val, pos, ok := evalExpr(filtered, 0, channelVars)
	if !ok {
		return 0, false
	}
	// Allow trailing EOF or consumed all tokens.
	if pos < len(filtered) &&
		filtered[pos].Kind != scanner.EOF {
		return 0, false
	}
	return val, true
}

// evalExpr parses additive expressions: term ((+|-) term)*
func evalExpr(
	tokens []scanner.Token,
	pos int,
	vars map[string]float64,
) (float64, int, bool) {
	left, pos, ok := evalTerm(tokens, pos, vars)
	if !ok {
		return 0, pos, false
	}

	for pos < len(tokens) {
		op, isOp := isAddSubOp(tokens[pos])
		if !isOp {
			break
		}
		pos++
		right, newPos, ok := evalTerm(tokens, pos, vars)
		if !ok {
			return 0, newPos, false
		}
		pos = newPos
		if op == "+" {
			left += right
		} else {
			left -= right
		}
	}

	return left, pos, true
}

// evalTerm parses multiplicative expressions:
// factor ((*|/) factor)*
func evalTerm(
	tokens []scanner.Token,
	pos int,
	vars map[string]float64,
) (float64, int, bool) {
	left, pos, ok := evalFactor(tokens, pos, vars)
	if !ok {
		return 0, pos, false
	}

	for pos < len(tokens) {
		tok := tokens[pos]
		if tok.Kind != scanner.Delim {
			break
		}
		if tok.Value != "*" && tok.Value != "/" {
			break
		}
		op := tok.Value
		pos++
		right, newPos, ok := evalFactor(tokens, pos, vars)
		if !ok {
			return 0, newPos, false
		}
		pos = newPos
		if op == "*" {
			left *= right
		} else {
			if right == 0 {
				return 0, pos, false
			}
			left /= right
		}
	}

	return left, pos, true
}

// evalFactor parses: number | percentage | ident | '(' expr ')'
// Also handles unary minus.
func evalFactor(
	tokens []scanner.Token,
	pos int,
	vars map[string]float64,
) (float64, int, bool) {
	if pos >= len(tokens) {
		return 0, pos, false
	}

	tok := tokens[pos]

	// Unary minus
	if isMinusToken(tok) {
		val, newPos, ok := evalFactor(tokens, pos+1, vars)
		if !ok {
			return 0, newPos, false
		}
		return -val, newPos, true
	}

	// Parenthesized expression
	if tok.Kind == scanner.ParenOpen {
		val, newPos, ok := evalExpr(tokens, pos+1, vars)
		if !ok {
			return 0, newPos, false
		}
		if newPos < len(tokens) &&
			tokens[newPos].Kind == scanner.ParenClose {
			return val, newPos + 1, true
		}
		return 0, newPos, false
	}

	// Number
	if tok.Kind == scanner.Number {
		v, err := strconv.ParseFloat(tok.Value, 64)
		if err != nil {
			return 0, pos, false
		}
		return v, pos + 1, true
	}

	// Percentage (treat as the numeric value, caller decides
	// meaning)
	if tok.Kind == scanner.Percentage {
		v, err := strconv.ParseFloat(tok.Value, 64)
		if err != nil {
			return 0, pos, false
		}
		return v, pos + 1, true
	}

	// Channel variable identifier
	if tok.Kind == scanner.Ident {
		if vars != nil {
			if v, ok := vars[tok.Value]; ok {
				return v, pos + 1, true
			}
		}
		return 0, pos, false
	}

	return 0, pos, false
}
