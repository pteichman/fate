package fate

import "sort"

// tokset maintains a set of tokens in a slice. Use Add to insert a
// token and use the slice itself to get the tokens.
type tokset []token

// Add inserts tok into this set, if not already present. It may
// return a new slice, so use its return value as the new set.
//
// Returns a bool signaling whether an add occurred.
func (t tokset) Add(tok token) (tokset, bool) {
	loc := t.search(tok)
	if loc == len(t) {
		return append(t, tok), true
	}

	if t[loc] == tok {
		return t, false
	}

	t = append(t, 0)
	copy(t[loc+1:], t[loc:])
	t[loc] = tok

	return t, true
}

func (t tokset) search(x token) int {
	// Add a fast path for empty arrays or brand new tokens.
	size := len(t)
	if size == 0 || x > t[size-1] {
		return size
	}

	return sort.Search(size, func(i int) bool { return t[i] >= x })
}
