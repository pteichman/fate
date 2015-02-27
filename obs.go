package fate

type obs struct {
	grams   map[token][]token
	bigrams map[bigram]tokset
}

func (o *obs) Observe(ctx bigram, tok2 token) {
	grams, bigrams := o.grams, o.bigrams

	set, ok := bigrams[ctx]
	if !ok && grams != nil {
		grams[ctx.tok0] = append(grams[ctx.tok0], ctx.tok1)
	}

	bigrams[ctx] = set.Add(tok2)
}

func (o *obs) Follow(tok token) []token {
	return o.grams[tok]
}

func (o *obs) FollowBigram(ctx bigram) []token {
	return o.bigrams[ctx]
}
