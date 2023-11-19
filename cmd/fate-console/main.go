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
	writeFilename := flag.String("w", "", "write all learned inputs to this file")
	maxlen := flag.Int("maxlen", 0, "maximum length for reply in UTF-8 chars")
	flag.Parse()

	model := fate.NewModel(fate.Config{})

	var (
		err       error
		writeFile *os.File
		toLearn   []string
	)
	if *writeFilename != "" {
		toLearn = append(toLearn, *writeFilename)
		writeFile, err = os.OpenFile(*writeFilename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			fmt.Printf("Error: opening log: %v\n", err)
			os.Exit(1)
		}
		defer writeFile.Close()
	}
	toLearn = append(toLearn, flag.Args()...)

	for _, f := range toLearn {
		err := learnFile(model, f)
		if err != nil {
			fmt.Printf("Error: %s\n", err)
			continue
		}
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
			if writeFile != nil {
				writeFile.WriteString(line + "\n")
			}
		}

		timeout := time.After(time.Second / 2)

		reply := model.Reply(line)
		for *maxlen > 0 && len(reply) > *maxlen {
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
