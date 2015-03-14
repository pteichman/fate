package fate

type obs1 map[token][]token

func (o obs1) Observe(tok0 token, tok1 token) {
	o[tok0] = append(o[tok0], tok1)
}

type obs2 map[bigram]tokset

func (o obs2) Observe(ctx bigram, tok2 token) bool {
	set, ok := o[ctx]

	if newset, added := set.Add(tok2); added {
		o[ctx] = newset
	}

	return ok
}
