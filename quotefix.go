package fate

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
)

// These functions automatically balance quotes/parens/etc in strings.
// Since fate's tokenizer splits only on spaces, replies often contain
// unmatched quotes or parentheses.

// QuoteFix automatically balances quotes/parens/etc in text strings.
func QuoteFix(s string) string {
	var qr []quoterune

	iter := newTokiter(s)
	for iter.Next() {
		tok := iter.Token()
		if candidate(tok) {
			qr = append(qr, quoterunes(tok)...)
		} else {
			qr = append(qr, literals(tok)...)
		}
	}

	return flatten(fixrev(fixfwd(qr)))
}

var isEmoticon = regexp.MustCompile(`:-*[\(\)]+`)

func candidate(tok string) bool {
	return !isEmoticon.MatchString(tok)
}

type quotetype int

const (
	literal quotetype = iota
	open
	close
)

func (qt quotetype) String() string {
	switch qt {
	case literal:
		return "literal"
	case open:
		return "open"
	case close:
		return "close"
	default:
		return "none"
	}
}

type quoterune struct {
	t quotetype
	r rune
}

func (qr quoterune) String() string {
	return fmt.Sprintf("{ %v %v }", qr.t, string(qr.r))
}

type tokiter struct {
	s   string
	tok string
}

func newTokiter(s string) *tokiter {
	return &tokiter{s, ""}
}

func (ti *tokiter) Next() bool {
	ti.s = ti.s[len(ti.tok):]
	if len(ti.s) == 0 {
		return false
	}

	notSpace := func(r rune) bool {
		return !unicode.IsSpace(r)
	}

	var end int

	first, _ := utf8.DecodeRuneInString(ti.s)
	if unicode.IsSpace(first) {
		end = strings.IndexFunc(ti.s, notSpace)
	} else {
		end = strings.IndexFunc(ti.s, unicode.IsSpace)
	}

	if end == -1 {
		end = len(ti.s)
	}

	ti.tok = ti.s[:end]

	return true
}

func (ti *tokiter) Token() string {
	return ti.tok
}

func literals(s string) []quoterune {
	var ret []quoterune
	for _, r := range s {
		ret = append(ret, quoterune{literal, r})
	}
	return ret
}

func quoterunes(s string) []quoterune {
	var ret []quoterune
	for i, r := range s {
		if r == '"' {
			dir := direction(s, i)
			ret = append(ret, quoterune{dir, r})
		} else if r == '(' || r == '{' || r == '[' || r == '“' {
			ret = append(ret, quoterune{open, r})
		} else if r == ')' || r == '}' || r == ']' || r == '”' {
			ret = append(ret, quoterune{close, r})
		} else {
			ret = append(ret, quoterune{literal, r})
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
		return open
	}

	return close
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

// fixfwd inserts close tokens for unmatched opens.
func fixfwd(tokens []quoterune) []quoterune {
	var stack []rune
	var prev rune

	var ret []quoterune
	for _, t := range tokens {
		if t.t == open {
			stack = append(stack, t.r)
		} else if len(stack) > 0 && t.t == close {
			stack, prev = pop(stack)
			if prev != mirror(t.r) {
				ret = append(ret, quoterune{close, mirror(prev)})
			}
		}

		ret = append(ret, t)
	}

	for len(stack) > 0 {
		stack, prev = pop(stack)
		ret = append(ret, quoterune{close, mirror(prev)})
	}

	return ret
}

// fixrev inserts open tokens for unmatched closes.
func fixrev(tokens []quoterune) []quoterune {
	var stack []rune
	var prev rune

	var ret []quoterune
	for i := len(tokens) - 1; i >= 0; i-- {
		t := tokens[i]
		if t.t == close {
			stack = append(stack, t.r)
		} else if len(stack) > 0 && t.t == open {
			stack, prev = pop(stack)
			if prev != mirror(t.r) {
				ret = append(ret, quoterune{open, mirror(prev)})
			}
		}

		ret = append(ret, t)
	}

	for len(stack) > 0 {
		stack, prev = pop(stack)
		ret = append(ret, quoterune{open, mirror(prev)})
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
	case '“':
		r = '”'
	case '”':
		r = '“'
	default:
	}

	return r
}
