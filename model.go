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

func (b bigram) reverse() bigram {
	return bigram{b.tok1, b.tok0}
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

// Config holds Model configuration data. An empty Config struct
// indicates the default values for each.
type Config struct {
	// Stemmer makes all tokens go through a normalization process
	// when created. Words that stem the same mean the same thing.
	Stemmer Stemmer
}

func (c Config) stemmerOrDefault() Stemmer {
	if c.Stemmer != nil {
		return c.Stemmer
	}

	return DefaultStemmer
}

// NewModel constructs an empty language model.
func NewModel(opts Config) *Model {
	return &Model{
		tokens: newSyndict(opts.stemmerOrDefault()),

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

	words := strings.Fields(text)
	if len(words) == 1 {
		// Refuse to learn single-word inputs.
		return
	}

	// ids: <S> <S> tokens in the input string </S> </S>
	var ids = []token{start, start}
	for _, f := range words {
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
	pivot := m.pickPivot(strings.Fields(text))

	fwdctx := bigram{tok0: pivot, tok1: m.choice(m.fwd1[pivot])}

	start, end := m.ends()
	fwd := m.follow(m.fwd2, fwdctx, end)
	rev := m.follow(m.rev2, fwdctx.reverse(), start)

	rev, fwd = cleanup(rev, fwd)

	return join(m.tokens, rev, fwd)
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

func (m *Model) follow(obs obs2, pos bigram, goal token) []token {
	if pos.tok1 == goal {
		if pos.tok0 == goal {
			return nil
		}
		return []token{pos.tok0}
	}

	ret := []token{pos.tok0, pos.tok1}

	for {
		toks := obs[pos]
		if len(toks) == 0 {
			log.Fatal("ran out of chain")
		}

		tok := m.choice(toks)
		if tok == goal {
			return ret
		}

		ret = append(ret, tok)
		pos.tok0, pos.tok1 = pos.tok1, tok
	}
}

func cleanup(rev, fwd []token) ([]token, []token) {
	// Clean up some artifacts of choosing start/end as a pivot.
	// TODO: don't do this
	if len(rev) == 1 {
		return nil, fwd[1:]
	} else if len(fwd) == 1 {
		return rev[1:], nil
	}

	return rev, fwd[2:]
}

func join(tokens *syndict, rev, fwd []token) string {
	buf := make([]byte, 0, joinsize(tokens, rev, fwd))

	reverse(rev)

	for _, tok := range rev {
		buf = append(buf, tokens.Word(tok)...)
		buf = append(buf, ' ')
	}

	for _, tok := range fwd {
		buf = append(buf, tokens.Word(tok)...)
		buf = append(buf, ' ')
	}

	return string(buf[:len(buf)-1])
}

func reverse(toks []token) {
	a, b := 0, len(toks)-1
	for a < b {
		toks[a], toks[b] = toks[b], toks[a]
		a++
		b--
	}
}

func joinsize(tokens *syndict, rev, fwd []token) int {
	// initialize count assuming a space between each word
	count := len(rev) + len(fwd)
	for _, tok := range rev {
		count += len(tokens.Word(tok))
	}

	for _, tok := range fwd {
		count += len(tokens.Word(tok))
	}

	return count
}

func (m *Model) choice(toks []token) token {
	return toks[m.rand.Intn(len(toks))]
}
