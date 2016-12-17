package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	//"unicode"
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
	a, err := Analyze(r.Body)
	if err != nil {
		// In principle it is possible that this is a client error,
		// but it would be a low-level (http-ish) error and probably
		// not an application-level error (e.g., trying to use deflate
		// transport encoding but sending garbage)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("could not read body\n"))
		return
	}
	buf, err := json.MarshalIndent(a, "", "  ")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("could not marshal response\n"))
	}
	// We do not respect Accept, only returning JSON
	w.Header().Set("Content-Type", "application/json")
	w.Write(append(buf, '\n'))
}

// Analyze processes the chat message from the reader, locating and
// noting references to entities.  Most malformed entity references
// are silently ignored.  Returns an error only if the
// given reader returns an error.
func Analyze(rd io.Reader) (*Analysis, error) {
	ret := &Analysis{}
	s := bufio.NewScanner(rd)
	s.Split(splitRef)
	for s.Scan() {
		log.Printf("*** %q", s.Text())
	}
	log.Printf("done scanning")

	return ret, nil
}

const scanWindow = 1024 // max URL length

var emoticonRe = regexp.MustCompile(`^\(([a-zA-Z]{1,15})\)`)
var mentionRe = regexp.MustCompile(`^@([a-zA-Z]+)`)
var urlRe = regexp.MustCompile(`#(?i)\b((?:[a-z][\w-]+:(?:/{1,3}|[a-z0-9%])|www\d{0,3}[.]|[a-z0-9.\-]+[.][a-z]{2,4}/)(?:[^\s()<>]+|\(([^\s()<>]+|(\([^\s()<>]+\)))*\))+(?:\(([^\s()<>]+|(\([^\s()<>]+\)))*\)|[^\s` + "`" + `!()\[\]{};:'".,<>?«»“”‘’]))#iS`)

func splitRef(data []byte, atEOF bool) (int, []byte, error) {
	// if atEOF is true, we have to find any token that is there; it's not
	// OK to let Scanner advance through data for us.  Unfortunately.
	alreadySkipped := 0

	for len(data) > 0 {
		log.Printf("hi (%d+) %q", alreadySkipped, data)
		// find the start of something potentially interesting
		k := bytes.IndexAny(data, "(h@")
		log.Printf("k=%d", k)
		if k < 0 {
			// there are no possible starts... consume everything
			// and return return no tokens
			return alreadySkipped + len(data), nil, nil
		}
		log.Printf("It is '%c'", data[k])

		switch data[k] {
		case '@':
			m := mentionRe.FindSubmatch(data[k:])
			if m != nil {
				log.Printf("MENTION %q", string(m[1]))
				return alreadySkipped + k + len(m[0]), m[1], nil
			}
			log.Printf("not a mention, remain %q", data[k+1:])
		}
		alreadySkipped += k + 1
		data = data[k+1:]
	}
	return alreadySkipped, nil, nil
}
