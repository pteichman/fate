// Package fate is a text generation library.
package fate

import (
	cryptorand "crypto/rand"
	"encoding/binary"
	"log"
	"math/rand"
	"strings"
)

type token uint32

type bigram struct {
	tok0, tok1 token
}

// Model is a trigram language model that can learn and respond to
// text.
type Model struct {
	tokens *syndict

	// We track (tok0 -> tok1) and (tok0 tok1 -> tok2) in
	// the forward direction, so we can efficiently choose
	// random tok0-containing contexts.
	fwd1 obs1
	fwd2 obs2

	// In the reverse direction, we only need to be able
	// to track (tok2 tok1 -> tok0).
	rev2 obs2

	rand *rand.Rand
}

// NewModel constructs an empty language model.
func NewModel() *Model {
	return &Model{
		tokens: newSyndict(strings.ToLower),

		fwd1: make(obs1),
		fwd2: make(obs2),

		rev2: make(obs2),

		rand: rand.New(randSource()),
	}
}

// randSource seeds a standard math/rand PRNG with a secure seed.
func randSource() rand.Source {
	var seed int64
	binary.Read(cryptorand.Reader, binary.LittleEndian, &seed)
	return rand.NewSource(seed)
}

func (m *Model) ends() (token, token) {
	return m.tokens.ID("<S>"), m.tokens.ID("</S>")
}

// Learn observes the text in a string and makes it available for
// later replies.
func (m *Model) Learn(text string) {
	start, end := m.ends()
	tokens := m.tokens

	// ids: <S> <S> tokens in the input string </S> </S>
	var ids = []token{start, start}
	for _, f := range strings.Fields(text) {
		ids = append(ids, tokens.ID(f))
	}
	ids = append(ids, end, end)

	var ctx bigram
	var tok2 token

	for i := 0; i < len(ids)-2; i++ {
		ctx.tok0, ctx.tok1, tok2 = ids[i], ids[i+1], ids[i+2]
		if !m.fwd2.Observe(ctx, tok2) {
			m.fwd1.Observe(ctx.tok0, ctx.tok1)
		}

		ctx.tok0, tok2 = tok2, ctx.tok0
		m.rev2.Observe(ctx, tok2)
	}
}

// Reply generates a reply string to str, given the current state of
// the language model.
func (m *Model) Reply(text string) string {
	start, end := m.ends()

	pivot := m.pickPivot(strings.Fields(text))

	// find a random context containing pivot
	fwdtoks := m.fwd1[pivot]

	fwdctx := bigram{tok0: pivot, tok1: m.choice(fwdtoks)}
	revctx := bigram{tok0: fwdctx.tok1, tok1: fwdctx.tok0}

	fwd := m.follow(m.fwd2, fwdctx, end)
	rev := m.follow(m.rev2, revctx, start)

	return m.join(rev, fwd[2:])
}

func (m *Model) pickPivot(words []string) token {
	var pivots []token
	for _, w := range words {
		pivots = append(pivots, m.tokens.Syns(w)...)
	}

	if len(pivots) > 0 {
		return pivots[m.rand.Intn(len(pivots))]
	}

	// No valid pivots, so babble.
	start, _ := m.ends()
	return start
}

func (m *Model) join(rev, fwd []token) string {
	var words = make([]string, 0, len(rev)+len(fwd))

	start, end := m.ends()
	add := func(tok token) {
		if tok != start && tok != end {
			words = append(words, m.tokens.Word(tok))
		}
	}

	for i := len(rev) - 1; i >= 0; i-- {
		add(rev[i])
	}

	for _, tok := range fwd {
		add(tok)
	}

	return strings.Join(words, " ")
}

func (m *Model) follow(obs obs2, pos bigram, end token) []token {
	ret := []token{pos.tok0, pos.tok1}

	if pos.tok1 == end {
		return ret
	}

	for {
		toks := obs[pos]
		if len(toks) == 0 {
			log.Fatal("ran out of chain")
		}

		tok := m.choice(toks)
		if tok == end {
			return ret
		}

		ret = append(ret, tok)
		pos.tok0, pos.tok1 = pos.tok1, tok
	}
}

func (m *Model) choice(toks []token) token {
	return toks[m.rand.Intn(len(toks))]
}
