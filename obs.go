package fate

type obs1 map[token][]token

func (o obs1) Observe(tok0 token, tok1 token) {
	o[tok0] = append(o[tok0], tok1)
}

type obs2 map[bigram]tokset

func (o obs2) Observe(ctx bigram, tok2 token) (had2, had3 bool) {
	set, had2 := o[ctx]

	if newset, had3 := set.Add(tok2); !had3 {
		o[ctx] = newset
		return had2, false
	}

	return had2, true
}
