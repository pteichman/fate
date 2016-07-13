package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path"
	"time"

	"github.com/peterh/liner"
	"github.com/pteichman/fate"
)

var historyFn = ".fate_console"

func main() {
	var (
		maxlen int
	)

	flag.IntVar(&maxlen, "maxlen", 0, "maximum length for reply in UTF-8 chars")
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

	console := liner.NewLiner()
	console.SetCtrlCAborts(true)
	defer console.Close()

	hist := path.Join(os.Getenv("HOME"), historyFn)
	if hist != historyFn {
		loadHistory(console, hist)
	}

loop:
	for {
		line, err := console.Prompt("> ")
		if err != nil {
			fmt.Println()
			break
		}

		if line != "" {
			console.AppendHistory(line)
		}

		timeout := time.After(time.Second / 2)

		reply := model.Reply(line)
		for maxlen > 0 && len(reply) > maxlen {
			reply = model.Reply(line)

			select {
			case <-timeout:
				fmt.Println("ERROR: timed out")
				continue loop
			default:
			}
		}

		fmt.Println(reply)
	}

	if hist != historyFn {
		saveHistory(console, hist)
	}
}

func learnFile(m *fate.Model, path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	s := bufio.NewScanner(f)
	for s.Scan() {
		m.Learn(s.Text())
	}

	return s.Err()
}

func loadHistory(console *liner.State, filename string) {
	f, err := os.Open(filename)
	if err != nil && !os.IsNotExist(err) {
		fmt.Printf("Reading %s: %s\n", filename, err)
		return
	}

	console.ReadHistory(f)
	f.Close()
}

func saveHistory(console *liner.State, filename string) {
	f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		fmt.Println(err)
		return
	}

	_, err = console.WriteHistory(f)
	if err != nil {
		fmt.Println(err)
	}
}
