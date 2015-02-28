package fate

// dict stores 1-grams, mapping between strings and an integer id.
type dict struct {
	words []string
	ids   map[string]token
}

func newDict() *dict {
	return &dict{ids: make(map[string]token)}
}

func (d *dict) Len() int {
	return len(d.words)
}

func (d *dict) ID(w string) token {
	if id, ok := d.ids[w]; ok {
		return id
	}

	id := token(len(d.words))
	d.words = append(d.words, w)
	d.ids[w] = id
	return id
}

// syn maintains lists of token synonyms.
type syn struct {
	// Index by token id.
	syns map[string]tokset
	key  func(string) string
}

func (s *syn) Add(word string, tok token) {
	key := s.key(word)
	s.syns[key] = s.syns[key].Add(tok)
}

func (s *syn) Get(word string) tokset {
	key := s.key(word)
	return s.syns[key]
}
