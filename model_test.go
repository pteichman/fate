package fate

import (
	"strings"
	"testing"
)

func TestModel_Reply(t *testing.T) {
	model := NewModel(Config{})

	text := "this is a test"

	model.Learn(text)
	reply := model.Reply(text)

	if reply != text {
		t.Errorf("Reply(this is a test) => %s, want %s", reply, text)
	}
}

// TestConflate ensures an unlearned token isn't in conflate()
// results, a learned one is regardless of how it's stemmed.
func TestModel_Conflate(t *testing.T) {
	model := NewModel(Config{})

	toks := model.conflate(strings.Fields("_"))
	if len(toks) != 0 {
		t.Errorf("conflate(_) => [%d]token, want [0]token", len(toks))
	}

	model.Learn("foo bar _ baz")

	toks = model.conflate(strings.Fields("_"))
	if len(toks) != 1 {
		t.Errorf("conflate(_) => [%d]token, want [1]token", len(toks))
	}
}

func TestModel_Babble(t *testing.T) {
	model := NewModel(Config{})

	text := "this is a test"

	model.Learn(text)

	for i := 0; i < 1000; i++ {
		reply := model.Reply("unknown")

		if reply != text {
			t.Fatalf("Reply(this is a test) => %s, want %s", reply, text)
		}

		if _, ok := model.tokens.CheckID("unknown"); ok {
			t.Fatalf("Reply(\"unknown\") registered token")
		}
	}
}

func TestModel_Duel(t *testing.T) {
	model := NewModel(Config{})

	model.Learn("this is a test")
	model.Learn("this is another test")

	for i := 0; i < 1000; i++ {
		reply := model.Reply("this")

		if reply != "this is a test" && reply != "this is another test" {
			t.Errorf("Reply(this is a test) => %s, want %s", reply, "this is (a|another) test")
		}
	}
}

func TestModel_Empty(t *testing.T) {
	// Make sure Model doesn't panic when empty.
	model := NewModel(Config{})

	reply := model.Reply("")
	if reply != "" {
		t.Errorf("Reply() => %s, want empty string", reply)
	}

	model.Learn("")

	reply = model.Reply("")
	if reply != "" {
		t.Errorf("Reply() => %s, want empty string", reply)
	}
}

// BenchmarkModel_Overhead checks the constant overhead of learning
// already-learned trigrams.
func BenchmarkModel_Overhead(b *testing.B) {
	quote := "On two occasions I have been asked, 'Pray, Mr. Babbage, if you put into the machine wrong figures, will the right answers come out?' I am not able rightly to apprehend the kind of confusion of ideas that could provoke such a question."

	m := NewModel(Config{})
	m.Learn(quote)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		m.Learn(quote)
	}
}
