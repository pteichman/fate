package fate

import (
	"strings"
	"unicode"

	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

// Stemmer normalizes a string to its stem.
type Stemmer interface {
	Stem(string) string
}

// DefaultStemmer makes reply inputs insensitive to case, accents, and
// punctuation.
var DefaultStemmer = &cleaner{}

type cleaner struct{}

func (c *cleaner) Stem(s string) string {
	tran := transform.Chain(norm.NFD, transform.RemoveFunc(isNonWord), norm.NFC)
	ret, _, err := transform.String(tran, strings.ToLower(s))
	if err != nil {
		return s
	}
	return ret
}

// isNonWord returns strippable Unicode characters: non-spacing marks
// and other punctuation.
func isNonWord(r rune) bool {
	return unicode.In(r, unicode.Mn, unicode.P)
}
