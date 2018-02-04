package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
)

func printHelp() {
	message := `Usage:

	lnchk [URL]
`
	fmt.Printf(message)
}

type link struct {
	URL          string        `json:"url"`
	Latency      time.Duration `json:"latency"`
	ResponseCode int           `json:"responseCode"`
	Error        error         `json:"error"`
}

func NewLink(url string, latency time.Duration, responseCode int, err error) *link {
	l := new(link)
	l.URL = url
	l.Latency = latency
	l.ResponseCode = responseCode
	l.Error = err
	return l
}

type summary struct {
	URL              string      `json:"url"`
	AvgLatency       float64     `json:"avgLatency"`
	ResponseCode     int         `json:"responseCode"`
	TotalLinks       int         `json:"totalLinks"`
	ResponsesPerCode map[int]int `json:"responsesPerCode"`
	Links            []link
}

func NewSummary(url string) *summary {
	s := new(summary)
	s.URL = url
	s.ResponsesPerCode = make(map[int]int)
	return s
}

func (s *summary) AddLink(l *link) {
	s.Links = append(s.Links, *l)
	s.TotalLinks += 1
	s.AvgLatency = s.AvgLatency + ((float64(l.Latency) - s.AvgLatency) / float64(s.TotalLinks))
	if _, ok := s.ResponsesPerCode[l.ResponseCode]; ok {
		s.ResponsesPerCode[l.ResponseCode]++
	} else {
		s.ResponsesPerCode[l.ResponseCode] = 1
	}
}

func ValidateArgs(args []string) error {
	givenArgs := len(args)

	if givenArgs == 1 {
		return errors.New("Missing URL")
	}

	if givenArgs > 2 {
		return fmt.Errorf("Got %d Arguments, expected 1", givenArgs-1)
	}

	return nil
}

func ParseLinkHref(pageURL *url.URL, href string) *url.URL {
	u, err := url.Parse(href)
	if err == nil {
		if u.Host == "" {
			u.Host = pageURL.Host
			if !strings.HasPrefix(u.Path, "/") {
				u.Path = path.Join(path.Dir(pageURL.Path), u.Path)
			}
		}

		if u.Scheme == "" {
			u.Scheme = pageURL.Scheme
		}
	}
	return u
}

func main() {

	err := ValidateArgs(os.Args)
	if err != nil {
		fmt.Printf("Error: %s\n\n", err.Error())
		printHelp()
		os.Exit(1)
	}

	pageURL, err := url.Parse(os.Args[1])
	if err != nil {
		fmt.Printf("Error parsing the given URL: %s\n", err.Error())
		os.Exit(1)
	}

	resp, err := http.Get(pageURL.String())
	if err != nil {
		fmt.Printf("Error getting the given URL: %s\n", err.Error())
		os.Exit(1)
	}
	defer resp.Body.Close()

	pageSummary := NewSummary(pageURL.String())
	pageSummary.ResponseCode = resp.StatusCode

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		fmt.Printf("Error parsing the document: %s\n", err.Error())
		os.Exit(1)
	}

	var n sync.WaitGroup
	doc.Find("a, link").Each(func(i int, s *goquery.Selection) {
		if href, ok := s.Attr("href"); ok {
			n.Add(1)
			go func(href string) {
				u := ParseLinkHref(pageURL, href)

				start := time.Now()
				r, e := http.Get(u.String())
				finish := time.Now()
				r.Body.Close()

				l := NewLink(u.String(), finish.Sub(start), r.StatusCode, e)
				pageSummary.AddLink(l)
				n.Done()
			}(href)
		}
	})
	n.Wait()

	json, _ := json.Marshal(pageSummary)
	fmt.Println(string(json))
}
