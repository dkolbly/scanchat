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
		given: "@donovan check out http://mock.it/fail (angry)",
		result: Analysis{
			Mentions:  []string{"donovan"},
			Emoticons: []string{"angry"},
			Links: []Link{
				Link{
					URL:   "http://mock.it/fail",
					Error: "mock failure",
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
