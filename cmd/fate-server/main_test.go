package main

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/pteichman/fate"
)

func NewServer(model *fate.Model) *httptest.Server {
	return httptest.NewServer(NewHandler(model))
}

func TestRoot(t *testing.T) {
	ts := NewServer(fate.NewModel(fate.Config{}))
	defer ts.Close()

	res, err := http.Get(ts.URL)
	if err != nil {
		t.Fatal(err)
	}

	if res.StatusCode != http.StatusNotFound {
		t.Fatalf("GET / -> %v, want %v", res.StatusCode, http.StatusNotFound)
	}
}

func TestLearn(t *testing.T) {
	ts := NewServer(fate.NewModel(fate.Config{}))
	defer ts.Close()

	args := url.Values{}
	args.Add("q", "foo bar baz")

	res, err := http.PostForm(ts.URL+"/learn", args)
	if err != nil {
		t.Fatal(err)
	}

	if res.StatusCode != http.StatusOK {
		t.Fatalf("POST /learn -> %v, want %v", res.StatusCode, http.StatusOK)
	}
}

func TestLearnGet(t *testing.T) {
	ts := NewServer(fate.NewModel(fate.Config{}))
	defer ts.Close()

	res, err := http.Get(ts.URL + "/learn")
	if err != nil {
		t.Fatal(err)
	}

	if res.StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("GET /learn -> %v, want %v", res.StatusCode, http.StatusMethodNotAllowed)
	}
}

func TestReply(t *testing.T) {
	model := fate.NewModel(fate.Config{})
	model.Learn("foo bar baz")

	ts := NewServer(model)
	defer ts.Close()

	req, err := http.NewRequest("GET", ts.URL+"/reply", nil)
	if err != nil {
		t.Fatal(err)
	}

	args := req.URL.Query()
	args.Add("q", "foo")
	req.URL.RawQuery = args.Encode()

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	if res.StatusCode != http.StatusOK {
		t.Fatalf("GET /reply -> %v, want %v", res.StatusCode, http.StatusOK)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	res.Body.Close()

	if string(body) != "foo bar baz" {
		t.Fatalf("GET /reply?q=foo -> %v, want %v", string(body), "foo bar baz")
	}
}

func TestReplyTimeout(t *testing.T) {
	model := fate.NewModel(fate.Config{})
	model.Learn("foo bar baz")

	ts := NewServer(model)
	defer ts.Close()

	req, err := http.NewRequest("GET", ts.URL+"/reply", nil)
	if err != nil {
		t.Fatal(err)
	}

	args := req.URL.Query()
	args.Add("q", "foo")
	args.Add("maxlen", "1")
	req.URL.RawQuery = args.Encode()

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	if res.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("GET /reply -> %v, want %v", res.StatusCode, http.StatusServiceUnavailable)
	}
}
