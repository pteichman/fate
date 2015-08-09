package fate

import (
	"math/rand"
	"sort"
)

// tokset maintains a set of tokens in a slice. Use Add to insert a
// token and use the slice itself to get the tokens.
type tokset struct {
	t1 []uint8
	t2 []uint16
	t3 []byte
	t4 []uint32
}

// Add inserts tok into this set if not already present.
//
// Returns a bool signaling whether the token was already in the set
// (similar logic to map lookups).
func (t *tokset) Add(tok token) bool {
	var found bool

	switch {
	case tok <= 0xFF:
		t.t1, found = add1(t.t1, tok)
		return found
	case tok <= 0xFFFF:
		t.t2, found = add2(t.t2, tok)
		return found
	case tok <= 0xFFFFFF:
		t.t3, found = add3(t.t3, tok)
		return found
	case tok <= 0xFFFFFFFF:
		t.t4, found = add4(t.t4, tok)
		return found
	}

	panic("oops")
}

func add1(buf []uint8, tok token) ([]byte, bool) {
	val := uint8(tok)

	size := len(buf)
	if size == 0 {
		return append(buf, val), false
	}

	loc := sort.Search(size, func(i int) bool { return buf[i] >= val })
	if loc < size && buf[loc] == val {
		return buf, true
	}

	buf = append(buf, 0)
	copy(buf[loc+1:], buf[loc:])
	buf[loc] = val

	return buf, false
}

func add2(buf []uint16, tok token) ([]uint16, bool) {
	val := uint16(tok)

	size := len(buf)
	if size == 0 {
		return append(buf, val), false
	}

	loc := sort.Search(size, func(i int) bool { return buf[i] >= val })
	if loc < len(buf) && buf[loc] == val {
		return buf, true
	}

	buf = append(buf, 0)
	copy(buf[loc+1:], buf[loc:])
	buf[loc] = val

	return buf, false
}

func add3(buf []byte, tok token) ([]byte, bool) {
	val := make([]byte, 3)
	put3(val, tok)

	size := len(buf) / 3
	if size == 0 {
		return append(buf, val...), false
	}

	loc := sort.Search(size, func(i int) bool { return unpack3(buf[3*i:]) >= tok })

	if 3*loc < len(buf) && unpack3(buf[3*loc:]) == tok {
		return buf, true
	}

	buf = append(buf, 0, 0, 0)
	copy(buf[3*(loc+1):], buf[3*loc:])
	copy(buf[3*loc:], val)

	return buf, false
}

func add4(buf []uint32, tok token) ([]uint32, bool) {
	val := uint32(tok)

	size := len(buf)
	if size == 0 {
		return append(buf, val), false
	}

	loc := sort.Search(size, func(i int) bool { return buf[i] >= val })
	if loc < len(buf) && buf[loc] == val {
		return buf, true
	}

	buf = append(buf, 0)
	copy(buf[loc+1:], buf[loc:])
	buf[loc] = val

	return buf, false
}

func (t *tokset) Len() int {
	if t == nil {
		return 0
	}

	return len(t.t1) + len(t.t2) + len(t.t3)/3 + len(t.t4)
}

func (t *tokset) Tokens() []token {
	if t == nil {
		return nil
	}

	var tokens []token

	for _, val := range t.t1 {
		tokens = append(tokens, token(val))
	}
	for _, val := range t.t2 {
		tokens = append(tokens, token(val))
	}
	for i := 0; i < len(t.t3); i += 3 {
		tokens = append(tokens, unpack3(t.t3[i:]))
	}
	for _, val := range t.t4 {
		tokens = append(tokens, token(val))
	}

	return tokens
}

func put3(buf []byte, tok token) {
	buf[0] = byte(tok)
	buf[1] = byte(tok >> 8)
	buf[2] = byte(tok >> 16)
}

func unpack3(buf []byte) token {
	return token(buf[0]) | token(buf[1])<<8 | token(buf[2])<<16
}

func (t tokset) Choice(r *rand.Rand) token {
	index := r.Intn(t.Len())

	switch {
	case index < len(t.t1):
		return token(t.t1[index])
	case index < len(t.t1)+len(t.t2):
		return token(t.t2[index-len(t.t1)])
	case index < len(t.t1)+len(t.t2)+len(t.t3):
		return unpack3(t.t3[index-len(t.t1)-len(t.t2):])
	case index < len(t.t1)+len(t.t2)+len(t.t3)+len(t.t4):
		return token(t.t4[index-len(t.t1)-len(t.t2)-len(t.t3)])
	}

	panic("oops")
}
