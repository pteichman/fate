package fate

type obs1 map[token][]token

func (o obs1) Observe(tok0 token, tok1 token) {
	o[tok0] = append(o[tok0], tok1)
}

type trigrams struct {
	fwd, rev obs2
}

func (t trigrams) Observe(tok0, tok1, tok2 token) (had2, had3 bool) {
	ctx := bigram{tok0, tok1}

	chain, had2 := t.fwd[ctx]
	if !had2 {
		chain = &tokset{}
		t.fwd[ctx] = chain
	}

	if had3 := chain.Add(tok2); !had3 {
		t.ObserveRev(tok2, tok1, tok0)
		return had2, false
	}

	return had2, had3
}

func (t trigrams) ObserveRev(tok0, tok1, tok2 token) {
	ctx := bigram{tok0, tok1}

	chain, had2 := t.rev[ctx]
	if !had2 {
		chain = &tokset{}
		t.rev[ctx] = chain
	}

	chain.Add(tok2)
}

func (t *trigrams) Fwd(ctx bigram) *tokset {
	return t.fwd[ctx]
}

func (t *trigrams) Rev(ctx bigram) *tokset {
	return t.rev[ctx]
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
