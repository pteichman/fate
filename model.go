// Package fate is a text generation library.
package fate

import (
	cryptorand "crypto/rand"
	"encoding/binary"
	"log"
	"math/rand"
	"strings"
	"sync"
	"unicode"
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

	lock *sync.Mutex
	rand *rand.Rand
}

// Config holds Model configuration data. An empty Config struct
// indicates the default values for each.
type Config struct {
	// Stemmer makes all tokens go through a normalization process
	// when created. Words that stem the same mean the same thing.
	Stemmer Stemmer
	Rand    rand.Source
}

func (c Config) stemmerOrDefault() Stemmer {
	if c.Stemmer != nil {
		return c.Stemmer
	}

	return DefaultStemmer
}

func (c Config) randOrDefault() rand.Source {
	if c.Rand != nil {
		return c.Rand
	}

	var seed int64
	binary.Read(cryptorand.Reader, binary.LittleEndian, &seed)
	return rand.NewSource(seed)
}

// NewModel constructs an empty language model.
func NewModel(opts Config) *Model {
	return &Model{
		tokens: newSyndict(opts.stemmerOrDefault()),

		fwd1: make(obs1),
		fwd2: make(obs2),

		rev2: make(obs2),

		lock: &sync.Mutex{},
		rand: rand.New(opts.randOrDefault()),
	}
}

func (m *Model) ends() (token, token) {
	return m.tokens.ID("<S>"), m.tokens.ID("</S>")
}

// Learn observes the text in a string and makes it available for
// later replies.
func (m *Model) Learn(text string) {
	m.lock.Lock()
	defer m.lock.Unlock()

	if !learnable(text) {
		// Refuse to learn single-word inputs.
		return
	}

	iter := m.newCtxiter(text)

	for iter.next() {
		m.observe(iter.trigram())
	}
}

type ctxiter struct {
	s    string
	dict *syndict
	end  token

	ctx bigram
	tok token
}

func (m *Model) newCtxiter(s string) *ctxiter {
	start, end := m.ends()

	return &ctxiter{
		s:    strings.TrimFunc(s, unicode.IsSpace),
		dict: m.tokens,
		end:  end,

		ctx: bigram{start, start},
		tok: start,
	}
}

func (ci *ctxiter) next() bool {
	s := ci.s

	fieldStart := -1
	fieldEnd := len(s)

	for i, rune := range s {
		if unicode.IsSpace(rune) {
			if fieldStart >= 0 {
				fieldEnd = i
				break
			}
		} else if fieldStart == -1 {
			fieldStart = i
		}
	}

	if fieldStart == -1 {
		// had no field; shift in the end token.
		ci.ctx = bigram{ci.ctx.tok1, ci.tok}
		ci.tok = ci.end
		return ci.ctx.tok0 != ci.end
	}

	// have a token from s[fieldStart:fieldEnd]
	ci.ctx = bigram{ci.ctx.tok1, ci.tok}
	ci.tok = ci.dict.ID(s[fieldStart:fieldEnd])
	ci.s = s[fieldEnd:]

	return true
}

func (ci *ctxiter) trigram() (bigram, token) {
	return ci.ctx, ci.tok
}

func learnable(s string) bool {
	n := 0
	inField := false
	for _, rune := range s {
		wasInField := inField
		inField = !unicode.IsSpace(rune)
		if inField && !wasInField {
			n++
		}

		if n > 1 {
			return true
		}
	}

	return false
}

func (m *Model) observe(ctx bigram, tok token) {
	// Observe the trigram: (tok0, tok1, tok2).
	old2, old3 := m.fwd2.Observe(ctx, tok)
	if !old2 {
		// If the bigram was new, observe that in fwd1.
		m.fwd1.Observe(ctx.tok0, ctx.tok1)
	}

	if !old3 {
		// If the trigram was new, observe that in rev2.
		ctx.tok0, tok = tok, ctx.tok0
		m.rev2.Observe(ctx, tok)
	}
}

// Reply generates a reply string to str, given the current state of
// the language model.
func (m *Model) Reply(text string) string {
	m.lock.Lock()
	tokens := m.conflate(strings.Fields(text))
	reply := join(m.tokens, m.replyTokens(tokens))
	m.lock.Unlock()

	return reply
}

func (m *Model) replyTokens(tokens []token) []token {
	pivot := m.pickPivot(tokens)

	fwdctx := bigram{tok0: pivot, tok1: m.choice(m.fwd1[pivot])}

	start, end := m.ends()

	var path []token

	// Compute the beginning of the sentence by walking from
	// fwdctx back to start.
	path = m.follow(path, m.rev2, fwdctx.reverse(), start)

	// Reverse what we have so far.
	reverse(path)

	// Append the initial context, tok0 and tok1. But tok0 only if
	// we weren't already at the start.
	if fwdctx.tok0 != start {
		path = append(path, fwdctx.tok0)
	}

	// And tok1 only if we weren't already at the end.
	if fwdctx.tok1 != end {
		path = append(path, fwdctx.tok1)

		// Compute the end of the sentence by walking forward
		// from fwdctx to end.
		path = m.follow(path, m.fwd2, fwdctx, end)
	}

	return path
}

func (m *Model) conflate(words []string) []token {
	var pivots []token
	for _, w := range words {
		syns := m.tokens.Syns(w)
		if tok, ok := m.tokens.CheckID(w); ok && !in(syns, tok) {
			pivots = append(pivots, tok)
		}

		pivots = append(pivots, syns...)
	}

	return pivots
}

func in(haystack []token, needle token) bool {
	for _, tok := range haystack {
		if tok == needle {
			return true
		}
	}

	return false
}

func (m *Model) pickPivot(tokens []token) token {
	if len(tokens) > 0 {
		return tokens[m.rand.Intn(len(tokens))]
	}

	// No valid pivots, so babble. Assume tokens 0 & 1 are start and end.
	return token(m.rand.Intn(m.tokens.Len()-2) + 2)
}

func (m *Model) follow(path []token, obs obs2, pos bigram, goal token) []token {
	for {
		toks := obs[pos]
		if len(toks) == 0 {
			log.Fatal("ran out of chain at", pos)
		}

		tok := toks.Choice(m.rand)
		if tok == goal {
			return path
		}

		path = append(path, tok)
		pos.tok0, pos.tok1 = pos.tok1, tok
	}
}

func join(tokens *syndict, path []token) string {
	if len(path) == 0 {
		return ""
	}

	buf := make([]byte, 0, joinsize(tokens, path))

	buf = append(buf, tokens.Word(path[0])...)
	for _, tok := range path[1:] {
		buf = append(buf, ' ')
		buf = append(buf, tokens.Word(tok)...)
	}

	return string(buf)
}

func reverse(toks []token) {
	a, b := 0, len(toks)-1
	for a < b {
		toks[a], toks[b] = toks[b], toks[a]
		a++
		b--
	}
}

func joinsize(tokens *syndict, path []token) int {
	// initialize count assuming a space between each word
	count := len(path) - 1
	for _, tok := range path {
		count += len(tokens.Word(tok))
	}

	return count
}

func (m *Model) choice(toks []token) token {
	return toks[m.rand.Intn(len(toks))]
}
