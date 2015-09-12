package fate

import (
	"strings"
	"unicode"
)

// words splits a string into words using unicode.IsSpace.
type words struct {
	str  string
	word string
}

func newWords(str string) *words {
	return &words{
		str: strings.TrimFunc(str, unicode.IsSpace),
	}
}

func (w *words) Next() bool {
	s := w.str

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
		return false
	}

	// have a word from s[fieldStart:fieldEnd]
	w.word = s[fieldStart:fieldEnd]
	w.str = s[fieldEnd:]

	return true
}

func (w *words) Word() string {
	return w.word
}
