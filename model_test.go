package fate

import (
	"bufio"
	"os"
	"testing"
)

func TestReply(t *testing.T) {
	model := NewModel()

	text := "this is a test"

	model.Learn(text)
	reply := model.Reply(text)

	if reply != text {
		t.Errorf("Reply(this is a test) => %s, want %s", reply, text)
	}
}

func TestBabble(t *testing.T) {
	model := NewModel()

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
	model := NewModel()

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
