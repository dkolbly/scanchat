package main

import (
	"bytes"
	"net/http"
	"regexp"
)

type Link struct {
	URL   string `json:"url"`
	Title string `json:"title,omitempty"`
	Error string `json:"error,omitempty"`
}

type Getter func(url string) (*http.Response, error)

type Analyzer struct {
	cur *Analysis
}

type Analysis struct {
	Mentions  []string `json:"mentions,omitempty"`
	Emoticons []string `json:"emoticons,omitempty"`
	Links     []Link   `json:"links,omitempty"`
}

// Analyze processes the chat message from the reader, locating and
// noting references to entities.  Most malformed entity references
// are silently ignored.
func Analyze(chat []byte, getter Getter) *Analysis {
	ret := &Analysis{}

	for len(chat) > 0 {
		k := bytes.IndexAny(chat, "(hH@")
		if k < 0 {
			break
		}
		if k > 0 {
			chat = chat[k:]
		}

		var remain []byte
		switch chat[0] {
		case '@':
			remain = ret.parseMention(chat)
		case '(':
			remain = ret.parseEmoticon(chat)
		case 'h', 'H':
			remain = ret.parseLink(chat, getter)
		}
		// if we parsed out a token,
		if remain == nil {
			chat = chat[1:]
		} else {
			chat = remain
		}
	}

	return ret
}

var mentionRe = regexp.MustCompile(`^@(\w+)`)
var emoticonRe = regexp.MustCompile(`^\(([\dA-Za-z]{1,15})\)`)

// parseMention parses out an @mention, if any, from a chat message
// starting at `@`, adding it to the Analysis.  Returns the remainder
// of the chat message (i.e., after the @mention), or nil if a valid
// mention was not found.
func (a *Analysis) parseMention(chat []byte) []byte {
	m := mentionRe.FindSubmatch(chat)
	if m == nil {
		return nil
	}

	a.Mentions = append(a.Mentions, string(m[1]))
	return chat[len(m[0]):]
}

// parseEmoticon parses out an (emoticon), if any, from a chat message
// starting at the `(`, adding it to the Analysis.  Returns the remainder
// of the chat message (i.e., after the (emoticon)), or nil if a valid
// emoticon was not found.
func (a *Analysis) parseEmoticon(chat []byte) []byte {
	m := emoticonRe.FindSubmatch(chat)
	if m == nil {
		return nil
	}

	a.Emoticons = append(a.Emoticons, string(m[1]))
	return chat[len(m[0]):]
}
