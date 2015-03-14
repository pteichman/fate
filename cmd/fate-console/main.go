package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path"

	"github.com/GeertJohan/go.linenoise"
	"github.com/pteichman/fate"
)

func main() {
	flag.Parse()

	model := fate.NewModel(fate.Config{})

	var learned bool
	for _, f := range flag.Args() {
		err := learnFile(model, f)
		if err != nil {
			fmt.Printf("Error: %s\n", err)
			continue
		}

		learned = true
	}

	if !learned {
		fmt.Println("Usage: fate-console <text files>")
		os.Exit(1)
	}

	chat(model)
}

func learnFile(m *fate.Model, path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}

	s := bufio.NewScanner(f)
	for s.Scan() {
		m.Learn(s.Text())
	}

	return s.Err()
}

func chat(m *fate.Model) {
	history := loadHistory()

	for {
		if err := chatOnce(m); err != nil {
			break
		}
	}

	if history != "" {
		saveHistory(history)
	}
}

func chatOnce(m *fate.Model) error {
	line, err := linenoise.Line("> ")
	if err != nil {
		return err
	}

	if line != "" {
		linenoise.AddHistory(line)
	}

	fmt.Println(m.Reply(line))

	return nil
}

func loadHistory() string {
	home := os.Getenv("HOME")
	if home == "" {
		return home
	}

	history := path.Join(home, ".fate_history")

	err := linenoise.LoadHistory(history)
	if err != nil {
		fmt.Println(err)
	}

	return history
}

func saveHistory(filename string) error {
	err := linenoise.SaveHistory(filename)
	if err != nil {
		fmt.Println(err)
	}

	return err
}
