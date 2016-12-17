package main

import (
	"io/ioutil"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
)

func main() {
	var port = flag.Int("http", 8080, "port to listen on")
	flag.Parse()

	http.Handle("/scan-message", &ChatParser{})
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *port), nil))
}

type ChatParser struct {
}

type Link struct {
	URL   string `json:"url"`
	Title string `json:"title,omitempty"`
	Error string `json:"error,omitempty"`
}

type Analysis struct {
	Mentions  []string `json:"mentions,omitempty"`
	Emoticons []string `json:"emoticons,omitempty"`
	Links     []Link   `json:"links,omitempty"`
}

func (s *ChatParser) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	input, err := ioutil.ReadAll(r.Body)
	if err != nil {
		// hmm, it might be possible that this is a client error
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("could not read body\n"))
		return
	}
		
	a := Analyze(string(input))
	
	buf, err := json.MarshalIndent(a, "", "  ")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("could not marshal response\n"))
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(append(buf, '\n'))
}

func Analyze(msg string) *Analysis {
	ret := &Analysis{}
	return ret
}
