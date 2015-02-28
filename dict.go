package fate

// dict stores 1-grams, mapping between strings and an integer id.
type dict struct {
	words []string
	ids   map[string]token
}

func newDict() *dict {
	return &dict{ids: make(map[string]token)}
}

func (d *dict) CheckID(w string) (token, bool) {
	tok, ok := d.ids[w]
	return tok, ok
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

func (d *dict) Word(tok token) string {
	return d.words[tok]
}

// syndict maintains lists of token synonyms.
type syndict struct {
	d *dict

	syns map[string]tokset
	key  func(string) string
}

func (s *syndict) CheckID(word string) (token, bool) {
	tok, ok := s.d.CheckID(word)
	return tok, ok
}

func (s *syndict) ID(word string) token {
	if tok, ok := s.d.CheckID(word); ok {
		return tok
	}

	tok := s.d.ID(word)
	key := s.key(word)
	s.syns[key] = s.syns[key].Add(tok)

	return tok
}

func (s *syndict) Syns(word string) tokset {
	key := s.key(word)
	return s.syns[key]
}

func (s *syndict) Word(tok token) string {
	return s.d.Word(tok)
}

func newSyndict(key func(string) string) *syndict {
	return &syndict{
		d:    newDict(),
		syns: make(map[string]tokset),
		key:  key,
	}
}
