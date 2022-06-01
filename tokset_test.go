package fate

import (
	"math/rand"
	"reflect"
	"testing"
)

func TestAdd(t *testing.T) {
	var tests = []struct {
		toks     []token
		expected *tokset
	}{
		{[]token{9, 7, 5, 3, 1}, toks(1, 3, 5, 7, 9)},
	}

	for _, tt := range tests {
		var ts tokset

		for _, tok := range tt.toks {
			ts.Add(tok)
		}

		if !reflect.DeepEqual(ts.Tokens(), tt.expected.Tokens()) {
			t.Errorf("Add(%v) -> %v, expected %v", tt.toks, ts, tt.expected)
		}
	}
}

func TestAddLarger(t *testing.T) {
	var tests = []struct {
		toks     []token
		expected *tokset
	}{
		{[]token{0, 1, 0xFF, 0xFF + 1, 0xFFFF, 0xFFFF + 1, 0xFFFFFF},
			toks(0, 1, 0xFF, 0xFF+1, 0xFFFF, 0xFFFF+1, 0xFFFFFF)},
	}

	for _, tt := range tests {
		var ts tokset

		for _, tok := range tt.toks {
			ts.Add(tok)
		}

		if !reflect.DeepEqual(ts.Tokens(), tt.expected.Tokens()) {
			t.Errorf("Add(%v) -> %v, expected %v", tt.toks, ts, tt.expected)
		}
	}
}

func BenchmarkToksetAdd(b *testing.B) {
	var ts tokset

	rnd := rand.New(rand.NewSource(0))

	for i := 0; i < b.N; i++ {
		ts.Add(token(rnd.Intn(100000)))
	}
}

func BenchmarkBitmapAdd(b *testing.B) {
	bm := NewBitmap()

	rnd := rand.New(rand.NewSource(0))

	for i := 0; i < b.N; i++ {
		bm.Add(token(rnd.Intn(100000)))
	}
}
