package main

import (
	"testing"
	"time"
)

func TestIndexNames(t *testing.T) {
	tc := []struct {
		prefix     string
		start, end string // in ISO8601
		expected   string
	}{
		{"a", "2017-05-27T07:00:00Z", "2017-05-27T07:00:00Z", "a-201721"},
		{"a", "2017-05-20T07:00:00Z", "2017-05-27T07:00:00Z", "a-201720,a-201721"},
		{"a", "2017-05-13T07:00:00Z", "2017-05-27T07:00:00Z", "a-201719,a-201720,a-201721"},
	}

	for _, c := range tc {
		s, _ := time.Parse(time.RFC3339, c.start)
		e, _ := time.Parse(time.RFC3339, c.end)
		r := indexNames(c.prefix, s, e)
		if r != c.expected {
			t.Fatalf("%s, but %s", c.expected, r)
		}
	}
}
