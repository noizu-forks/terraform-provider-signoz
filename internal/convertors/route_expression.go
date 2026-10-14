package conv

import "strings"

// CanonicalRouteExpression normalizes a route-policy alert expression to the
// canonical operator/whitespace form used by the provider: `AND`/`OR` word
// operators become `&&`/`||`, and whitespace runs outside string literals
// collapse to single spaces. The server echoes expressions back verbatim, so
// state written by older configs (or the UI) drifts from canonical configs
// and every plan flags a text-only change; both the write (Expand) and read
// (Flatten) paths canonicalize to keep plans stable.
func CanonicalRouteExpression(expr string) string {
	var b strings.Builder
	b.Grow(len(expr))

	inSingle, inDouble, inBacktick := false, false, false
	pendingSpace := false

	for i := 0; i < len(expr); i++ {
		c := expr[i]

		if inSingle || inDouble || inBacktick {
			b.WriteByte(c)

			switch {
			case c == '\'' && inSingle:
				inSingle = false
			case c == '"' && inDouble:
				inDouble = false
			case c == '`' && inBacktick:
				inBacktick = false
			}

			continue
		}

		switch c {
		case '\'', '"', '`':
			if pendingSpace {
				b.WriteByte(' ')
				pendingSpace = false
			}

			switch c {
			case '\'':
				inSingle = true
			case '"':
				inDouble = true
			case '`':
				inBacktick = true
			}

			b.WriteByte(c)
		case ' ', '\t', '\n', '\r':
			pendingSpace = true
		default:
			if pendingSpace {
				b.WriteByte(' ')
				pendingSpace = false
			}

			switch {
			case matchExprWord(expr, i, "AND"):
				b.WriteString("&&")
				i += 2
			case matchExprWord(expr, i, "OR"):
				b.WriteString("||")
				i += 1
			default:
				b.WriteByte(c)
			}
		}
	}

	if pendingSpace {
		b.WriteByte(' ')
	}

	return strings.TrimRight(b.String(), " ")
}

// matchExprWord reports whether the exact word sits at index i with
// identifier-character boundaries on both sides (so label names like
// `ANDROID` or `ORDER` never rewrite).
func matchExprWord(expr string, i int, word string) bool {
	end := i + len(word)

	if end > len(expr) || expr[i:end] != word {
		return false
	}

	if i > 0 && isExprIdentChar(expr[i-1]) {
		return false
	}

	if end < len(expr) && isExprIdentChar(expr[end]) {
		return false
	}

	return true
}

func isExprIdentChar(c byte) bool {
	switch {
	case c >= 'a' && c <= 'z', c >= 'A' && c <= 'Z', c >= '0' && c <= '9':
		return true
	case c == '_' || c == '.' || c == '-':
		return true
	default:
		return false
	}
}
