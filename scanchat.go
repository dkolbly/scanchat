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

func (s *ChatParser) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	chat, err := ioutil.ReadAll(r.Body)
	if err != nil {
		// In principle it is possible that this is a client error,
		// but it would be a low-level (http-ish) error and probably
		// not an application-level error (e.g., trying to use deflate
		// transport encoding but sending garbage)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("could not read body\n"))
		return
	}
	a := Analyze(chat)
	buf, err := json.MarshalIndent(a, "", "  ")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("could not marshal response\n"))
	}
	// We do not respect Accept, only returning JSON
	w.Header().Set("Content-Type", "application/json")
	w.Write(append(buf, '\n'))
}

