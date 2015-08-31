package fate

import "unicode"

type ctxiter struct {
	s    string
	dict *syndict
	end  token

	ctx bigram
	tok token
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
	ci.tok = ci.dict.ID(s[fieldStart:fieldEnd])
	ci.s = s[fieldEnd:]

	return true
}

func (ci *ctxiter) trigram() (bigram, token) {
	return ci.ctx, ci.tok
}
