package fate

import "testing"

func TestWords(t *testing.T) {
	var tests = []struct {
		str      string
		expected []string
	}{
		{"", []string{}},
		{"one", []string{"one"}},
		{"this is a test", []string{"this", "is", "a", "test"}},
		{"  this is a test  ", []string{"this", "is", "a", "test"}},
	}

	for _, tt := range tests {
		var words []string

		iter := newWords(tt.str)
		for iter.Next() {
			words = append(words, iter.Word())
		}

		if !StrsEqual(words, tt.expected) {
			t.Errorf("Words(%v) -> %v, expected %v", tt.str, words, tt.expected)
		}
	}
}

func StrsEqual(a []string, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	for i := 0; i < len(a); i++ {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}
