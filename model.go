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
	tokens *dict

	fwd *obs
	rev *obs

	rand *rand.Rand
}

// NewModel constructs an empty language model.
func NewModel() *Model {
	return &Model{
		tokens: newDict(),

		// We track (tok0 -> tok1) and (tok0 tok1 -> tok2) in
		// the forward direction, so we can efficiently choose
		// random tok0-containing contexts.
		fwd: &obs{
			grams:   make(map[token][]token),
			bigrams: make(map[bigram]tokset),
		},

		// In the reverse direction, we only need to be able
		// to track (tok2 tok1 -> tok0), so grams is not
		// allocated.
		rev: &obs{
			bigrams: make(map[bigram]tokset),
		},

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
		m.fwd.Observe(ctx, tok2)

		ctx.tok0, tok2 = tok2, ctx.tok0
		m.rev.Observe(ctx, tok2)
	}
}

// Reply generates a reply string to str, given the current state of
// the language model.
func (m *Model) Reply(text string) string {
	start, end := m.ends()

	toks := strings.Fields(text)

	pivot := toks[m.rand.Intn(len(toks))]
	pid := m.tokens.ID(pivot)

	// find a random context containing pid
	fwdtoks := m.fwd.Follow(pid)
	if len(fwdtoks) == 0 {
		// not yet.
		log.Fatal("don't know how to babble!")
	}

	fwdctx := bigram{tok0: pid, tok1: m.choice(fwdtoks)}
	revctx := bigram{tok0: fwdctx.tok1, tok1: fwdctx.tok0}

	fwd := m.follow(m.fwd, fwdctx, end)
	rev := m.follow(m.rev, revctx, start)

	if revctx.tok0 == end {
		// join() usually strips the sentence end token from
		// the end of the fwd chain, but rev will have an
		// extra if it came from the pivot context.
		rev = rev[1:]
	}

	return m.join(rev, fwd[2:])
}

func (m *Model) join(rev, fwd []token) string {
	var words = make([]string, 0, len(rev)+len(fwd))

	for i := len(rev) - 1; i >= 0; i-- {
		words = append(words, m.tokens.words[rev[i]])
	}

	for _, tok := range fwd {
		words = append(words, m.tokens.words[tok])
	}

	return strings.Join(words, " ")
}

func (m *Model) follow(obs *obs, pos bigram, end token) []token {
	ret := []token{pos.tok0, pos.tok1}

	for {
		toks := obs.FollowBigram(pos)
		if len(toks) == 0 {
			// not yet.
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
