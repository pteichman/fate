package fate

import (
	"math/rand"
	"reflect"
	"testing"
)

func TestAdd(t *testing.T) {
	var tests = []struct {
		toks     []token
		expected tokset
	}{
		{[]token{9, 7, 5, 3, 1}, tokset{1, 3, 5, 7, 9}},
	}

	for _, tt := range tests {
		var ts tokset

		for _, tok := range tt.toks {
			ts, _ = ts.Add(tok)
		}

		if !reflect.DeepEqual(ts, tt.expected) {
			t.Errorf("Add(%v) -> %v, expected %v", tt.toks, ts, tt.expected)
		}
	}
}

func BenchmarkToksetAdd(b *testing.B) {
	var ts tokset

	rnd := rand.New(rand.NewSource(0))

	for i := 0; i < b.N; i++ {
		ts, _ = ts.Add(token(rnd.Intn(100000)))
	}
}
