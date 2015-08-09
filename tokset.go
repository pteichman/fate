package fate

import (
	"math/rand"
	"sort"
)

// tokset maintains a set of tokens as a sorted slice of integers.
//
// 1-byte tokens (<= 0xFF) are in buf[0:t2]
// 2-byte tokens (<= 0xFFFF) are in buf[t2:t3]
// 3-byte tokens (<= 0xFFFFFF) are in buf[t3:]
//
// They're stored little-endian. Adds are O(log N). Choosing a random
// token in the set is O(1).
//
// tokens greater than 0xFFFFFF are not currently supported. This is
// enough token space to handle the Web 1T corpus.
type tokset struct {
	buf []byte

	// count of 2-byte tokens, count of 1-byte tokens
	c2 uint16
	c1 uint8
}

func (t *tokset) Add(tok token) bool {
	size := len(t.buf)

	switch {
	case tok <= 0xFF:
		if size == 0 {
			t.buf = append(t.buf, byte(tok))
			t.c1++
			return false
		}

		return t.add1(tok)
	case tok <= 0xFFFF:
		if size == 0 {
			t.buf = append(t.buf, byte(tok), byte(tok>>8))
			t.c2++
			return false
		}

		return t.add2(tok)
	case tok <= 0xFFFFFF:
		if size == 0 {
			t.buf = append(t.buf, byte(tok), byte(tok>>8), byte(tok>>16))
			return false
		}

		return t.add3(tok)
	}

	panic("oops")
}

func (t *tokset) span1() []byte {
	return t.buf[0:t.c1]
}

func (t *tokset) span2() []byte {
	return t.buf[int(t.c1) : int(t.c1)+2*int(t.c2)]
}

func (t *tokset) span3() []byte {
	return t.buf[int(t.c1)+2*int(t.c2):]
}

func (t *tokset) add1(tok token) bool {
	span := t.span1()
	loc := sort.Search(len(span), func(i int) bool {
		return token(span[i]) >= tok
	})

	if loc < len(span) && token(span[loc]) == tok {
		return true
	}

	t.buf = append(t.buf, 0)
	copy(t.buf[loc+1:], t.buf[loc:])
	t.buf[loc] = byte(tok)

	t.c1++

	return false
}

func (t *tokset) add2(tok token) bool {
	span := t.span2()
	idx := sort.Search(len(span)/2, func(i int) bool {
		return unpack2(span[2*i:]) >= tok
	})

	if idx < len(span)/2 && unpack2(span[2*idx:]) == tok {
		return true
	}

	t.buf = append(t.buf, 0, 0)

	loc := int(t.c1) + 2*idx
	copy(t.buf[loc+2:], t.buf[loc:])
	put2(t.buf[loc:], tok)

	t.c2++

	return false
}
func (t *tokset) add3(tok token) bool {
	span := t.span3()
	idx := sort.Search(len(span)/3, func(i int) bool {
		return unpack3(span[3*i:]) >= tok
	})

	if idx < len(span)/3 && unpack3(span[3*idx:]) == tok {
		return true
	}

	t.buf = append(t.buf, 0, 0, 0)

	loc := int(t.c1) + 2*int(t.c2) + 3*idx
	copy(t.buf[loc+3:], t.buf[loc:])
	put3(t.buf[loc:], tok)

	return false
}

func (t *tokset) Len() int {
	if t == nil {
		return 0
	}

	return int(t.c1) + int(t.c2) + len(t.span3())/3
}

func (t *tokset) Tokens() []token {
	if t == nil {
		return nil
	}

	var tokens []token
	for _, val := range t.span1() {
		tokens = append(tokens, token(val))
	}

	span2 := t.span2()
	for i := 0; i < len(span2); i += 2 {
		tokens = append(tokens, unpack2(span2[i:]))
	}

	span3 := t.span3()
	for i := 0; i < len(span3); i += 3 {
		tokens = append(tokens, unpack3(span3[i:]))
	}

	return tokens
}

func put2(buf []byte, tok token) {
	buf[0] = byte(tok)
	buf[1] = byte(tok >> 8)
}

func put3(buf []byte, tok token) {
	buf[0] = byte(tok)
	buf[1] = byte(tok >> 8)
	buf[2] = byte(tok >> 16)
}

func unpack2(buf []byte) token {
	return token(buf[0]) | token(buf[1])<<8
}

func unpack3(buf []byte) token {
	return token(buf[0]) | token(buf[1])<<8 | token(buf[2])<<16
}

func (t tokset) Choice(r *rand.Rand) token {
	index := r.Intn(t.Len())

	switch {
	case index < int(t.c1):
		return token(t.buf[index])
	case index < int(t.c1)+int(t.c2):
		span := t.span2()
		return unpack2(span[2*(index-int(t.c1)):])
	case index < t.Len():
		span := t.span3()
		return unpack3(span[3*(index-(int(t.c2)+int(t.c1))):])
	}

	panic("oops")
}
