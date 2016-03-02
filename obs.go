package fate

type bigrams map[token]*tokset

func (b bigrams) Observe(tok0 token, tok1 token) {
	ctx, ok := b[tok0]
	if !ok {
		ctx = &tokset{}
		b[tok0] = ctx
	}
	ctx.Add(tok1)
}

type fwdrev struct {
	fwd tokset
	rev tokset
}

type trigrams map[bigram]*fwdrev

func (t trigrams) Observe(tok0, tok1, tok2, tok3 token) (had2 bool) {
	ctx := bigram{tok1, tok2}

	chain, had2 := t[ctx]
	if !had2 {
		chain = &fwdrev{}
		t[ctx] = chain
	}

	chain.fwd.Add(tok3)
	chain.rev.Add(tok0)

	return had2
}

func (t trigrams) Fwd(ctx bigram) *tokset {
	return &(t[ctx].fwd)
}

func (t trigrams) Rev(ctx bigram) *tokset {
	return &(t[ctx].rev)
}
