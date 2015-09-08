package fate

import (
	"strings"
	"unicode"
)

// words splits a string into words using unicode.IsSpace.
type words struct {
	str  string
	word string
}

func newWords(str string) *words {
	return &words{
		str: strings.TrimFunc(str, unicode.IsSpace),
	}
}

func (w *words) Next() bool {
	s := w.str

	fieldStart := -1
	fieldEnd := len(s)

	for i, rune := range s {
		if unicode.IsSpace(rune) {
			if fieldStart >= 0 {
				fieldEnd = i
				break
			}
		} else if fieldStart == -1 {
			fieldStart = i
		}
	}

	if fieldStart == -1 {
		return false
	}

	// have a word from s[fieldStart:fieldEnd]
	w.word = s[fieldStart:fieldEnd]
	w.str = s[fieldEnd:]

	return true
}

func (w *words) Word() string {
	return w.word
}

// ctxiter splits a string into trigrams using unicode.IsSpace. It's
// optimized for minimal allocations. The trigrams are provided as
// (bigram, token) pairs.
type ctxiter struct {
	s        string
	word2tok func(string) token
	end      token

	ctx bigram
	tok token
}

func newCtxiter(str string, start, end token, word2tok func(string) token) *ctxiter {
	return &ctxiter{
		s:        strings.TrimFunc(str, unicode.IsSpace),
		word2tok: word2tok,
		end:      end,

		ctx: bigram{start, start},
		tok: start,
	}
}

func (ci *ctxiter) next() bool {
	s := ci.s

	fieldStart := -1
	fieldEnd := len(s)

	for i, rune := range s {
		if unicode.IsSpace(rune) {
			if fieldStart >= 0 {
				fieldEnd = i
				break
			}
		} else if fieldStart == -1 {
			fieldStart = i
		}
	}

	if fieldStart == -1 {
		// had no field; shift in the end token.
		ci.ctx = bigram{ci.ctx.tok1, ci.tok}
		ci.tok = ci.end
		return ci.ctx.tok0 != ci.end
	}

	// have a token from s[fieldStart:fieldEnd]
	ci.ctx = bigram{ci.ctx.tok1, ci.tok}
	ci.tok = ci.word2tok(s[fieldStart:fieldEnd])
	ci.s = s[fieldEnd:]

	return true
}

func (ci *ctxiter) trigram() (bigram, token) {
	return ci.ctx, ci.tok
}
