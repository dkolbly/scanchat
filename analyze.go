package main

import (
	"regexp"
	"bufio"
	"bytes"
	"log"
)

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

// Analyze processes the chat message from the reader, locating and
// noting references to entities.  Most malformed entity references
// are silently ignored.
func Analyze(chat []byte) *Analysis {
	ret := &Analysis{}
	s := bufio.NewScanner(bytes.NewReader(chat))
	s.Split(splitRef)
	for s.Scan() {
		log.Printf("*** %q", s.Text())
	}
	log.Printf("done scanning")

	return ret
}

const scanWindow = 1024 // max URL length

var emoticonRe = regexp.MustCompile(`^\(([a-zA-Z]{1,15})\)`)
var mentionRe = regexp.MustCompile(`^@([a-zA-Z]+)`)

// The @gruber v2 URL regex from https://mathiasbynens.be/demo/url-regex
// seems to be a good compromise between complexity and completeness (leaning towards
// fail "safe" where safe is defined as recognizing URLs).  This turns out to be
// a dark art, and most likely each organization should have a standard one that we
// should use here.
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
