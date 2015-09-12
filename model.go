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
	bi bigrams

	tri trigrams

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

		bi:  make(bigrams),
		tri: make(trigrams),

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

	start, end := m.startTok, m.endTok

	// Maintain a four-token sliding window. This allows us to learn
	// both the forward and reverse directions from the {tok1, tok2}
	// bigram at the same time.
	var (
		tok0 token = start
		tok1 token = start
		tok2 token = start
		tok3 token = 0
	)

	iter := newWords(text)

	m.lock.Lock()
	for iter.Next() {
		tok3 = m.tokens.ID(iter.Word())
		m.observe(tok0, tok1, tok2, tok3)
		tok0, tok1, tok2 = tok1, tok2, tok3
	}

	// Have: tok0=foo tok1=bar tok2=baz
	// Want: foo bar baz </S>
	//       bar baz </S> </S>
	//       baz </S> </S> </S>

	m.observe(tok0, tok1, tok2, end)
	m.observe(tok1, tok2, end, end)
	m.observe(tok2, end, end, end)

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

func (m *Model) observe(tok0, tok1, tok2, tok3 token) {
	// Observe the trigram: (tok0, tok1, tok2).
	if !m.tri.Observe(tok0, tok1, tok2, tok3) {
		m.bi.Observe(tok1, tok2)
	}
}

// Reply generates a reply string to str, given the current state of
// the language model. If no text has been learned, returns an empty
// string.
func (m *Model) Reply(text string) string {
	m.lock.RLock()
	defer m.lock.RUnlock()

	if m.tokens.Len() <= 2 {
		return ""
	}

	tokens := m.conflate(strings.Fields(text))
	reply := join(m.tokens, m.replyTokens(tokens, &prng{m.rand.Next()}))

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

	fwdctx := bigram{tok0: pivot, tok1: m.bi[pivot].Choice(intn)}

	start, end := m.startTok, m.endTok

	var path []token

	// Compute the beginning of the sentence by walking from
	// fwdctx back to start.
	path = m.followrev(path, m.tri, fwdctx, start)

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
		path = m.followfwd(path, m.tri, fwdctx, end)
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

func (m *Model) followfwd(path []token, tri trigrams, pos bigram, goal token) []token {
	for {
		toks := tri.Fwd(pos)
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

func (m *Model) followrev(path []token, tri trigrams, pos bigram, goal token) []token {
	for {
		toks := tri.Rev(pos)
		if toks.Len() == 0 {
			log.Fatal("ran out of chain at", pos)
		}

		tok := toks.Choice(m.rand)
		if tok == goal {
			return path
		}

		path = append(path, tok)
		pos.tok0, pos.tok1 = tok, pos.tok0
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
