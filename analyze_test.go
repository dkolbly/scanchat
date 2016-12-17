package main

import (
	"reflect"
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
}

func TestAnalyze(t *testing.T) {
	for _, c := range testCases {
		a := Analyze([]byte(c.given))
		if !reflect.DeepEqual(&c.result, a) {
			t.Errorf("Given %q, got %#v when expected %#v", c.given, a, &c.result)
		}
	}
}
