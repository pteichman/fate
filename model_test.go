package fate

import (
	"bufio"
	"os"
	"strings"
	"testing"
)

func TestReply(t *testing.T) {
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
func TestConflate(t *testing.T) {
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

func TestBabble(t *testing.T) {
	model := NewModel(Config{})

	text := "this is a test"

	model.Learn(text)

	for i := 0; i < 1000; i++ {
		reply := model.Reply("unknown")

		if reply != text {
			t.Errorf("Reply(this is a test) => %s, want %s", reply, text)
		}

		if _, ok := model.tokens.CheckID("unknown"); ok {
			t.Errorf("Reply(\"unknown\") registered token")
		}
	}
}

func TestDuel(t *testing.T) {
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

func learnFile(m *Model, filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	s := bufio.NewScanner(file)
	for s.Scan() {
		m.Learn(s.Text())
	}

	return s.Err()
}

var quote = "On two occasions I have been asked, 'Pray, Mr. Babbage, if you put into the machine wrong figures, will the right answers come out?' I am not able rightly to apprehend the kind of confusion of ideas that could provoke such a question."

// BenchmarkOverhead checks the constant overhead of learning
// already-learned trigrams.
func BenchmarkOverhead(b *testing.B) {
	m := NewModel(Config{})

	m.Learn(quote)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		m.Learn(quote)
	}
}
