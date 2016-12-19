package main

import (
	"bytes"
	"golang.org/x/net/html"
	"io"
	"log"
	"net/http"
	"regexp"
)

// urlRe is a modified form(*) of the @gruber v2 URL regex from
// https://mathiasbynens.be/demo/url-regex, which seems to be a good
// compromise between complexity and completeness (leaning towards
// fail "safe" where safe is defined as recognizing URLs).
//
// This turns out to be a dark art, and most likely each organization
// should have a standard one that we should use here.
//
// Also, note that this will successfully match only single balanced
// parentheticals which in this context is a desirable property
// because of the likelihood that someone writing something like "Hey
// @bob did you see the thing? (you know, http://bit.ly/thing)" does
// not intend the ')' to be part of the URL.  However, it is
// limited(**) in capability
//
// (* the modifications are to limit it to explicit http[s], and adapt
// it to our operating environment by anchoring it and, of course,
// expressing it as a Go string)
//
// (** this SO answer, my all-time favorite, explains one reason that
// it is limited, although in the context of HTML parsing and not URL
// parsing: http://bit.ly/1hY5QfK)
//                               ^ see, I just did it
//
var urlRe = regexp.MustCompile(`^(?i)\b((?:https?:(?:/{1,3}|[a-z0-9%]))(?:[^\s()<>]+|\(([^\s()<>]+|(\([^\s()<>]+\)))*\))+(?:\(([^\s()<>]+|(\([^\s()<>]+\)))*\)|[^\s` + "`" + `!()\[\]{};:'".,<>?«»“”‘’]))`)

// parseLink parses out an URL link, if any, from a chat message
// starting at the `h`, adding it to the Analysis.  Returns the remainder
// of the chat message (i.e., after the url), or nil if a valid
// url was not found.
func (a *Analysis) parseLink(chat []byte, getter Getter) []byte {
	m := urlRe.FindSubmatch(chat)
	if m == nil {
		return nil
	}

	link := Link{
		URL: string(m[0]),
	}
	title, err := getTitleForURL(getter, link.URL)
	if err != nil {
		link.Error = err.Error()
	} else {
		link.Title = title
	}

	a.Links = append(a.Links, link)
	return chat[len(m[0]):]
}

// ErrNotSuccessful represents a response other than 200 OK while
// attempting to retrieve a referenced page
type ErrNotSuccessful struct {
	Status string
}

// Error implements the error interface.  Returns the status
// string, which may be something like "404 Not found" or
// "500 Internal server error"
func (ns ErrNotSuccessful) Error() string {
	return ns.Status
}

// getTitleForURL dereferences the given URL using the supplied
// getter.  Since this is an internet access, some caching strategy
// is probably called for (see Caching Considerations in the README)
//
// Also, note that this is taking network action based on user-supplied
// content; see Security Considerations in the README
func getTitleForURL(getter Getter, url string) (string, error) {
	resp, err := getter(url)
	if err != nil {
		log.Printf("Retrieval of %s failed: %s", url, err)
		return "", err
	}
	log.Printf("Retrieval success: %s", resp.Status)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", ErrNotSuccessful{resp.Status}
	}

	return extractTitle(resp.Body)
}

// extractTitle reads the body, presumed to be HTML, and pulls out the
// content of the <title> element.  The underlying library,
// golang.org/x/net/html, seems to handle odd cases (like malformed
// HTML) pretty well and, in fact, appears to return the title text in
// a single TextToken -- but it's not clear if that's part of the API or
// just happens to be true for the current implementation, so we do
// support multiple TextTokens within the <title>.
func extractTitle(body io.Reader) (string, error) {
	rd := html.NewTokenizer(body)

	inTitle := false
	depth := 0
	var titleAccum bytes.Buffer

	for {
		tok := rd.Next()
		switch tok {
		case html.ErrorToken:
			return "", rd.Err()
		case html.StartTagToken:
			if inTitle {
				depth++
			} else {
				tag, _ := rd.TagName()
				if string(tag) == "title" {
					inTitle = true
					depth = 0
				}
			}
		case html.EndTagToken:
			if inTitle {
				if depth == 0 {
					return titleAccum.String(), nil
				}
				depth--
			}
		case html.TextToken:
			if inTitle {
				titleAccum.Write(rd.Text())
			}
		}
	}
}
