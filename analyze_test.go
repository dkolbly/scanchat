package main

import (
	"errors"
	"io/ioutil"
	"net/http"
	"reflect"
	"strings"
	"testing"
)

type simpleCase struct {
	given  string
	result Analysis
}

var testCases = []simpleCase{
	{
		given: "Nothing here",
	},
	{
		given: "@alice @bob",
		result: Analysis{
			Mentions: []string{"alice", "bob"},
		},
	},
	{
		given: "Fun (happy) stuff",
		result: Analysis{
			Emoticons: []string{"happy"},
		},
	},
	{
		given: "@donovan you around? (hungry)",
		result: Analysis{
			Mentions:  []string{"donovan"},
			Emoticons: []string{"hungry"},
		},
	},
	{
		given: "@donovan check out http://mock.it/good (smile)",
		result: Analysis{
			Mentions:  []string{"donovan"},
			Emoticons: []string{"smile"},
			Links: []Link{
				Link{
					URL:   "http://mock.it/good",
					Title: "Good stuff",
				},
			},
		},
	},
	{
		given: "@donovan I'm (sad) (see http://mock.it/fail)",
		result: Analysis{
			Mentions:  []string{"donovan"},
			Emoticons: []string{"sad"},
			Links: []Link{
				Link{
					URL:   "http://mock.it/fail",
					Error: "mock failure",
				},
			},
		},
	},
	{
		given: "see http://mock.it/entities",
		result: Analysis{
			Links: []Link{
				Link{
					URL:   "http://mock.it/entities",
					Title: "Does x<y?",
				},
			},
		},
	},
	{
		given: "see http://mock.it/subelem",
		result: Analysis{
			Links: []Link{
				Link{
					URL:   "http://mock.it/subelem",
					Title: "Cat <i>blue</i>",
				},
			},
		},
	},
}

func mockURLGetter(url string) (ret *http.Response, err error) {

	ret = &http.Response{
		Status:     "200 OK",
		StatusCode: 200,
		Header:     http.Header{},
	}

	var body string

	switch url {
	case "http://mock.it/fail": // for testing a low-level URL deref error
		return nil, errors.New("mock failure")

	case "http://mock.it/good": // for testing the happy path
		body = "<html><title>Good stuff</title><body>yay</body>"

	case "http://mock.it/entities": // for testing entity refs in the title
		body = "<html><title>Does x&lt;y&#x3f;</title><body>maybe</body>"

	case "http://mock.it/subelem":
		// This is bad HTML, but Chrome at least treats it as escaped,
		// as does golang.org/x/net/html (maybe not a coincidence, eh)
		body = "<html><title>Cat <i>blue</i></title><body>Blue</body>"
	}
	ret.Body = ioutil.NopCloser(strings.NewReader(body))
	return
}

func TestAnalyze(t *testing.T) {
	for _, c := range testCases {
		a := Analyze([]byte(c.given), mockURLGetter)
		if !reflect.DeepEqual(&c.result, a) {
			t.Errorf("Given %q, got %#v when expected %#v", c.given, a, &c.result)
		}
	}
}
