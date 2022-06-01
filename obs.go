package fate

type bigrams map[token]*Bitmap

func (b bigrams) Observe(tok0 token, tok1 token) {
	ctx, ok := b[tok0]
	if !ok {
		ctx = NewBitmap()
		b[tok0] = ctx
		stats.Add("TokenLearned", 1)
	}
	ctx.Add(tok1)
}

type fwdrev struct {
	fwd *Bitmap
	rev *Bitmap
}

type trigrams map[bigram]*fwdrev

func (t trigrams) Observe(tok0, tok1, tok2, tok3 token) (had2 bool) {
	ctx := bigram{tok1, tok2}

	chain, had2 := t[ctx]
	if !had2 {
		chain = &fwdrev{
			fwd: NewBitmap(),
			rev: NewBitmap(),
		}
		t[ctx] = chain
		stats.Add("BigramLearned", 1)
	}

	if !chain.fwd.Add(tok3) {
		stats.Add("TrigramLearned", 1)
	}

	chain.rev.Add(tok0)

	return had2
}

func (t trigrams) Fwd(ctx bigram) *Bitmap {
	return t[ctx].fwd
}

func (t trigrams) Rev(ctx bigram) *Bitmap {
	return t[ctx].rev
}
