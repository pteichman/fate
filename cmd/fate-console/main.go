package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path"

	"github.com/peterh/liner"
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
	line := liner.NewLiner()
	defer line.Close()

	history := loadHistory(line)

	for {
		if err := chatOnce(m, line); err != nil {
			break
		}
	}

	if history != "" {
		saveHistory(history, line)
	}
}

func chatOnce(m *fate.Model, console *liner.State) error {
	line, err := console.Prompt("> ")
	if err != nil {
		return err
	}

	if line != "" {
		console.AppendHistory(line)
	}

	fmt.Println(m.Reply(line))

	return nil
}

func loadHistory(line *liner.State) string {
	home := os.Getenv("HOME")
	if home == "" {
		return home
	}

	history := path.Join(home, ".fate_history")

	fd, err := os.Open(history)
	if err != nil {
		fmt.Println(err)
		return history
	}

	line.ReadHistory(fd)
	fd.Close()

	return history
}

func saveHistory(filename string, console *liner.State) error {
	f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}

	_, err = console.WriteHistory(f)
	if err != nil {
		fmt.Println(err)
	}

	return err
}
