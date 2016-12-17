package main

import (
	"regexp"
)

// The @gruber v2 URL regex from https://mathiasbynens.be/demo/url-regex
// seems to be a good compromise between complexity and completeness (leaning towards
// fail "safe" where safe is defined as recognizing URLs).  This turns out to be
// a dark art, and most likely each organization should have a standard one that we
// should use here.
var urlRe = regexp.MustCompile(`#(?i)\b((?:[a-z][\w-]+:(?:/{1,3}|[a-z0-9%])|www\d{0,3}[.]|[a-z0-9.\-]+[.][a-z]{2,4}/)(?:[^\s()<>]+|\(([^\s()<>]+|(\([^\s()<>]+\)))*\))+(?:\(([^\s()<>]+|(\([^\s()<>]+\)))*\)|[^\s` + "`" + `!()\[\]{};:'".,<>?«»“”‘’]))#iS`)

// parseLink parses out an URL link, if any, from a chat message
// starting at the `h`, adding it to the Analysis.  Returns the remainder
// of the chat message (i.e., after the url), or nil if a valid
// url was not found.
func (a *Analysis) parseLink(chat []byte) []byte {
	m := urlRe.FindSubmatch(chat)
	if m == nil {
		return nil
	}

	link := Link{
		URL: string(m[0]),
	}

	a.Links = append(a.Links, link)
	return chat[len(m[0]):]
}
