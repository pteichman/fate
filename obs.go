package fate

type obs1 map[token][]token

func (o obs1) Observe(tok0 token, tok1 token) {
	o[tok0] = append(o[tok0], tok1)
}

type fwdrev struct {
	fwd *tokset
	rev *tokset
}

type trigrams map[bigram]*fwdrev

func (t trigrams) Observe(tok0, tok1, tok2, tok3 token) (had2 bool) {
	ctx := bigram{tok1, tok2}

	chain, had2 := t[ctx]
	if !had2 {
		chain = &fwdrev{&tokset{}, &tokset{}}
		t[ctx] = chain
	}

	chain.fwd.Add(tok3)
	chain.rev.Add(tok0)

	return had2
}

func (t trigrams) Fwd(ctx bigram) *tokset {
	return t[ctx].fwd
}

func (t trigrams) Rev(ctx bigram) *tokset {
	return t[ctx].rev
}

type obs2 map[bigram]*tokset

func (o obs2) Observe(ctx bigram, tok2 token) (had2, had3 bool) {
	set, had2 := o[ctx]
	if !had2 {
		set = &tokset{}
		o[ctx] = set
	}

	if had3 := set.Add(tok2); !had3 {
		return had2, false
	}

	return had2, true
}
