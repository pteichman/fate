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
	tokens   *syndict
	startTok token
	endTok   token

	// We track (tok0 -> tok1) and (tok0 tok1 -> tok2) in the
	// forward direction, so we can efficiently choose random
	// tok0-containing contexts. In the reverse direction, we
	// only need to be able to track (tok2 tok1 -> tok0).
	fwd1 obs1

	tri *trigrams

	lock *sync.RWMutex
	rand *prng
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
	seed := opts.randOrDefault().Int63()
	tokens := newSyndict(opts.stemmerOrDefault())

	return &Model{
		tokens:   tokens,
		startTok: tokens.ID("<S>"),
		endTok:   tokens.ID("</S>"),

		fwd1: make(obs1),
		tri:  &trigrams{make(obs2), make(obs2)},

		lock: &sync.RWMutex{},
		rand: &prng{uint64(seed)},
	}
}

// Learn observes the text in a string and makes it available for
// later replies.
func (m *Model) Learn(text string) {
	if !learnable(text) {
		// Refuse to learn single-word inputs.
		return
	}

	m.lock.Lock()
	iter := newCtxiter(text, m.startTok, m.endTok, m.tokens.ID)
	for iter.next() {
		m.observe(iter.trigram())
	}
	m.lock.Unlock()
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
	old2, _ := m.tri.Observe(ctx, tok)
	if !old2 {
		// If the bigram was new, observe that in fwd1.
		m.fwd1.Observe(ctx.tok0, ctx.tok1)
	}
}

// Reply generates a reply string to str, given the current state of
// the language model. If no text has been learned, returns an empty
// string.
func (m *Model) Reply(text string) string {
	m.lock.RLock()
	if m.tokens.Len() <= 2 {
		return ""
	}

	tokens := m.conflate(strings.Fields(text))
	reply := join(m.tokens, m.replyTokens(tokens, &prng{m.rand.Next()}))
	m.lock.RUnlock()

	return reply
}

func (m *Model) replyTokens(tokens []token, intn Intn) []token {
	var pivot token
	if len(tokens) > 0 {
		pivot = choice(tokens, intn)
	} else {
		// Babble. Assume tokens 0 & 1 are start and end.
		pivot = token(intn.Intn(m.tokens.Len()-2) + 2)
	}

	fwdctx := bigram{tok0: pivot, tok1: choice(m.fwd1[pivot], intn)}

	start, end := m.startTok, m.endTok

	var path []token

	// Compute the beginning of the sentence by walking from
	// fwdctx back to start.
	path = m.follow(path, m.tri.Rev, fwdctx.reverse(), start)

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
		path = m.follow(path, m.tri.Fwd, fwdctx, end)
	}

	return path
}

func (m *Model) conflate(words []string) []token {
	var pivots = make([]token, 0, len(words))
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

func (m *Model) follow(path []token, next func(bigram) *tokset, pos bigram, goal token) []token {
	for {
		toks := next(pos)
		if toks.Len() == 0 {
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

func choice(toks []token, intn Intn) token {
	return toks[intn.Intn(len(toks))]
}
