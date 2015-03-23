package fate

import "sort"

// tokset maintains a set of tokens in a slice. Use Add to insert a
// token and use the slice itself to get the tokens.
type tokset []token

// Add inserts tok into this set, if not already present. It may
// return a new slice, so use its return value as the new set.
//
// Returns a bool signaling whether the token was already in the set
// (similar logic to map lookups).
func (t tokset) Add(tok token) (tokset, bool) {
	size := len(t)

	// Fast path for empty sets or brand new tokens.
	if size == 0 || tok > t[size-1] {
		return append(t, tok), false
	}

	loc := sort.Search(size, func(i int) bool { return t[i] >= tok })
	if t[loc] == tok {
		return t, true
	}

	t = append(t, 0)
	copy(t[loc+1:], t[loc:])
	t[loc] = tok

	return t, false
}
