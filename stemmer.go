package fate

import "strings"

// Stemmer normalizes a string to its stem.
type Stemmer interface {
	Stem(string) string
}

type stemFunc func(string) string

func (s stemFunc) Stem(str string) string {
	return s(str)
}

// DefaultStemmer is a stemmer that lowercases its tokens, making
// replies case-insensitive.
var DefaultStemmer = stemFunc(strings.ToLower)
