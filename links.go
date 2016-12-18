package main

import (
	"bytes"
	"golang.org/x/net/html"
	"io"
	"log"
	"net/http"
	"regexp"
)

// The @gruber v2 URL regex from https://mathiasbynens.be/demo/url-regex
// seems to be a good compromise between complexity and completeness (leaning towards
// fail "safe" where safe is defined as recognizing URLs).  This turns out to be
// a dark art, and most likely each organization should have a standard one that we
// should use here.
var urlRe = regexp.MustCompile(`(?i)\b((?:[a-z][\w-]+:(?:/{1,3}|[a-z0-9%])|www\d{0,3}[.]|[a-z0-9.\-]+[.][a-z]{2,4}/)(?:[^\s()<>]+|\(([^\s()<>]+|(\([^\s()<>]+\)))*\))+(?:\(([^\s()<>]+|(\([^\s()<>]+\)))*\)|[^\s` + "`" + `!()\[\]{};:'".,<>?«»“”‘’]))`)

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

func (ns ErrNotSuccessful) Error() string {
	return ns.Status
}

// getTitleForURL dereferences the given URL using the supplied
// getter.  Since this is an internet access, one hopes that the
// getter is caching results (perhaps using a caching proxy configured
// using HTTP_PROXY).  We should also instrument this to figure out if
// it'd be worth building our own cache of parsed out titles, to save
// us the work of actually parsing the HTML at all.  memcache would
// be a pretty good backing store for a title cache, because we can
// set expiration times so we can respect content that changes over time
// (redis would also work pretty well; there may be other solutions,
// but something like groupcache, while being nice for scaling out
// and avoiding the thundering herd, doesn't support expiration)
//
// See also Security Considerations in the README
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
// golang.org/x/net/html, seems to handle weird things pretty well and
// in fact appears to return the title text in a single TextToken, but
// it's not clear if that's part of the API or just happens to be true
// for the current implementation, so we do process multiple
// TextTokens within a <title>.
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
