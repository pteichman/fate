package fate

// These functions automatically balance quotes/parens/etc in strings.
// Since fate's tokenizer splits only on spaces, replies often contain
// unmatched quotes or parentheses.

func QuoteFix(s string) string {
	return flatten(fixrev(fixfwd(quoterunes(s))))
}

type quotetype int

const (
	Literal quotetype = iota
	Open
	Close
)

type quoterune struct {
	t quotetype
	r rune
}

func quoterunes(s string) []quoterune {
	var ret []quoterune
	for i, r := range s {
		if r == '"' {
			dir := direction(s, i)
			ret = append(ret, quoterune{dir, r})
		} else if r == '(' || r == '{' || r == '[' {
			ret = append(ret, quoterune{Open, r})
		} else if r == ')' || r == '}' || r == ']' {
			ret = append(ret, quoterune{Close, r})
		} else {
			ret = append(ret, quoterune{Literal, r})
		}
	}

	return ret
}

func direction(s string, pos int) quotetype {
	var (
		start = pos
		end   = pos
	)

	for start > 0 {
		if s[start] == ' ' {
			break
		}
		start--
	}

	for end < len(s) {
		if s[end] == ' ' {
			break
		}
		end++
	}

	if (pos - start) < (end - pos) {
		return Open
	}

	return Close
}

func pop(runes []rune) ([]rune, rune) {
	last := len(runes) - 1
	return runes[:last], runes[last]
}

func flatten(tokens []quoterune) string {
	var ret []rune
	for _, t := range tokens {
		ret = append(ret, t.r)
	}
	return string(ret)
}

// fixfwd inserts Close tokens for unmatched Opens.
func fixfwd(tokens []quoterune) []quoterune {
	var stack []rune
	var prev rune

	var ret []quoterune
	for _, t := range tokens {
		if t.t == Open {
			stack = append(stack, t.r)
		} else if len(stack) > 0 && t.t == Close {
			stack, prev = pop(stack)
			if prev != mirror(t.r) {
				ret = append(ret, quoterune{Close, mirror(t.r)})
			}
		}

		ret = append(ret, t)
	}

	for len(stack) > 0 {
		stack, prev = pop(stack)
		ret = append(ret, quoterune{Close, mirror(prev)})
	}

	return ret
}

// fixrev inserts Open tokens for unmatched Closes.
func fixrev(tokens []quoterune) []quoterune {
	var stack []rune
	var prev rune

	var ret []quoterune
	for i := len(tokens) - 1; i >= 0; i-- {
		t := tokens[i]
		if t.t == Close {
			stack = append(stack, t.r)
		} else if len(stack) > 0 && t.t == Open {
			stack, prev = pop(stack)
			if prev != mirror(t.r) {
				ret = append(ret, quoterune{Open, mirror(t.r)})
			}
		}

		ret = append(ret, t)
	}

	for len(stack) > 0 {
		stack, prev = pop(stack)
		ret = append(ret, quoterune{Open, mirror(prev)})
	}

	reverserunes(ret)

	return ret
}

func reverserunes(v []quoterune) {
	a, b := 0, len(v)-1
	for a < b {
		v[a], v[b] = v[b], v[a]
		a++
		b--
	}
}

func mirror(r rune) rune {
	switch r {
	case '(':
		r = ')'
	case ')':
		r = '('
	case '{':
		r = '}'
	case '}':
		r = '{'
	case '[':
		r = ']'
	case ']':
		r = '['
	case '"':
		r = '"'
	default:
	}

	return r
}
