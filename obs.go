package fate

type obs1 map[token][]token

func (o obs1) Observe(tok0 token, tok1 token) {
	o[tok0] = append(o[tok0], tok1)
}

type obs2 map[bigram]tokset

func (o obs2) Observe(ctx bigram, tok2 token) (old2, old3 bool) {
	set, old2 := o[ctx]

	if newset, old3 := set.Add(tok2); !old3 {
		o[ctx] = newset
		return old2, false
	}

	return old2, true
}
