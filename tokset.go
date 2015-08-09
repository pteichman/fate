package fate

import (
	"math/rand"
	"sort"
)

// tokset maintains a set of tokens in a slice. Use Add to insert a
// token and use the slice itself to get the tokens.
type tokset struct {
	t []token
}

// Add inserts tok into this set, if not already present. It may
// return a new slice, so use its return value as the new set.
//
// Returns a bool signaling whether the token was already in the set
// (similar logic to map lookups).
func (t *tokset) Add(tok token) bool {
	size := len(t.t)

	// Fast path for empty sets or brand new tokens.
	if size == 0 || tok > t.t[size-1] {
		t.t = append(t.t, tok)
		return false
	}

	loc := sort.Search(size, func(i int) bool { return t.t[i] >= tok })
	if t.t[loc] == tok {
		return true
	}

	t.t = append(t.t, 0)
	copy(t.t[loc+1:], t.t[loc:])
	t.t[loc] = tok

	return false
}

func (t *tokset) Len() int {
	if t == nil {
		return 0
	}

	return len(t.t)
}

func (t *tokset) Tokens() []token {
	if t == nil {
		return nil
	}

	return t.t
}

func (t tokset) Choice(r *rand.Rand) token {
	return t.t[r.Intn(t.Len())]
}
