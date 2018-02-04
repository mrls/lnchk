package main

import (
	"net/url"
	"reflect"
	"testing"
)

func TestValidateArgs(t *testing.T) {
	tests := []struct {
		in   []string
		want string
	}{
		{[]string{"lnchk"}, "Missing URL"},
		{[]string{"lnchk", "http://example.com", "http://foo.com"}, "Got 2 Arguments, expected 1"},
	}

	for _, test := range tests {
		err := ValidateArgs(test.in)
		if err != nil && err.Error() != test.want {
			t.Errorf("Given: %s\nwant: %s\ngot: %s", test.in, test.want, err.Error())
		}
	}

	err := ValidateArgs([]string{"lnchk", "http://example.com"})
	if err != nil {
		t.Error("Error occurred but wasn expecting any")
	}

}

func expectSummary(t *testing.T, s summary, links int, latency float64, responses map[int]int) {
	if s.TotalLinks != links {
		t.Errorf("Summary.TotalLinks, got: %d, want: %d.", s.TotalLinks, links)
	}

	if s.AvgLatency != latency {
		t.Errorf("Summary.AvgLatency, got: %f, want: %f.", s.AvgLatency, latency)
	}

	if !reflect.DeepEqual(s.ResponsesPerCode, responses) {
		t.Errorf("Summary.ResponsesPerCode, got: %#v, want: %#v.", s.ResponsesPerCode, responses)
	}
}

func TestAddLink(t *testing.T) {
	s := NewSummary("http://example.com")

	l := NewLink("http://example.com/about", 10, 200, nil)
	s.AddLink(l)

	expectSummary(t, *s, 1, 10, map[int]int{200: 1})

	l = NewLink("http://example.com/foo", 20, 200, nil)
	s.AddLink(l)

	expectSummary(t, *s, 2, 15, map[int]int{200: 2})

	l = NewLink("http://example.com/not-found", 30, 404, nil)
	s.AddLink(l)

	expectSummary(t, *s, 3, 20, map[int]int{200: 2, 404: 1})
}

func TestParseLinkHref(t *testing.T) {
	p1, _ := url.Parse("http://example.com")
	p2, _ := url.Parse("http://example.com/foo/")
	p3, _ := url.Parse("http://example.com/foo/bar.html")
	tests := []struct {
		pageURL *url.URL
		path    string
		want    string
	}{
		{p1, "about", "http://example.com/about"},
		{p2, "bar", "http://example.com/foo/bar"},
		{p3, "baz", "http://example.com/foo/baz"},
		{p2, "/baz", "http://example.com/baz"},
		{p2, "//foo.com", "http://foo.com"},
	}

	for _, test := range tests {
		url := ParseLinkHref(test.pageURL, test.path)
		if url.String() != test.want {
			t.Errorf("Given: %s\nwant: %s\ngot: %s", test.path, test.want, url.String())
		}
	}
}