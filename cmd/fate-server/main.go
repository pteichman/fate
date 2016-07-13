package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/pteichman/fate"
)

func main() {
	flag.Parse()

	if flag.NArg() == 0 {
		fmt.Println("Usage: fate-server <text files>")
		os.Exit(1)
	}

	model := fate.NewModel(fate.Config{})
	for _, f := range flag.Args() {
		err := learnFile(model, f)
		if err != nil {
			log.Printf("Learning %s: %s\n", f, err)
			continue
		}
	}

	http.Handle("/", NewHandler(model))

	srv := &http.Server{
		Addr:         ":8080",
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	log.Fatal(srv.ListenAndServe())
}

func NewHandler(m *fate.Model) http.Handler {
	h := handler{model: m}

	mux := http.NewServeMux()
	mux.HandleFunc("/learn", h.learn)
	mux.HandleFunc("/reply", h.reply)
	return mux
}

type handler struct {
	model *fate.Model
}

func (h handler) reply(w http.ResponseWriter, req *http.Request) {
	timeout := time.After(time.Second)

	q := req.FormValue("q")
	maxlen := parseint(req.FormValue("maxlen"))

	reply := h.model.Reply(q)
	for maxlen > 0 && len(reply) > maxlen {
		select {
		case <-timeout:
			http.Error(w, "Request timed out", http.StatusServiceUnavailable)
			return
		default:
		}

		reply = h.model.Reply(q)
	}

	w.Write([]byte(reply))
}

func (h handler) learn(w http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
	}

	q := req.FormValue("q")
	h.model.Learn(q)
}

func parseint(s string) int {
	v, err := strconv.ParseInt(s, 10, 32)
	if err != nil {
		v = 0
	}
	return int(v)
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
