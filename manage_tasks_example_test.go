package flowmatic_test

import (
	"fmt"
	"io"
	"maps"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing/fstest"

	"github.com/earthboundkid/flowmatic/v2"
)

func ExampleManager() {
	// Example site to crawl with recursive links
	srv := httptest.NewServer(http.FileServerFS(fstest.MapFS{
		"index.html": &fstest.MapFile{
			Data: []byte("/a.html"),
		},
		"a.html": &fstest.MapFile{
			Data: []byte("/b1.html\n/b2.html"),
		},
		"b1.html": &fstest.MapFile{
			Data: []byte("/c.html"),
		},
		"b2.html": &fstest.MapFile{
			Data: []byte("/c.html"),
		},
		"c.html": &fstest.MapFile{
			Data: []byte("/\n/x.html"),
		},
	}))
	defer srv.Close()
	cl := srv.Client()

	// Task fetches a page and extracts the URLs
	task := func(u string) ([]string, error) {
		res, err := cl.Get(srv.URL + u)
		if err != nil {
			return nil, err
		}
		defer res.Body.Close()
		if res.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("bad response: %q", u)
		}
		body, err := io.ReadAll(res.Body)
		if err != nil {
			return nil, err
		}

		return strings.Split(string(body), "\n"), nil
	}

	// results tracks which urls a url links to
	results := map[string][]string{}
	// tried tracks how many times a url has been queued to be fetched
	tried := map[string]int{}

	// Manage the tasks with as many workers as GOMAXPROCS
	m := flowmatic.Manage(flowmatic.MaxProcs, task)
	m.Queue("/")
	for range m.Exec() {
		req, urls, err := m.Result()
		if err != nil {
			// If there's a problem fetching a page, try three times
			if tried[req] < 3 {
				tried[req]++
				m.Queue(req)
			}
			continue
		}
		results[req] = urls
		for _, u := range urls {
			if tried[u] == 0 {
				m.Queue(u)
				tried[u]++
			}
		}
	}

	for _, key := range slices.Sorted(maps.Keys(results)) {
		fmt.Println(key, "links to:")
		for _, v := range results[key] {
			fmt.Println("- ", v)
		}
	}

	// Output:
	// / links to:
	// -  /a.html
	// /a.html links to:
	// -  /b1.html
	// -  /b2.html
	// /b1.html links to:
	// -  /c.html
	// /b2.html links to:
	// -  /c.html
	// /c.html links to:
	// -  /
	// -  /x.html
}

func ExampleManager_Work() {
	// Example site to crawl with recursive links
	srv := httptest.NewServer(http.FileServerFS(fstest.MapFS{
		"index.html": &fstest.MapFile{
			Data: []byte("/a.html"),
		},
		"a.html": &fstest.MapFile{
			Data: []byte("/b1.html\n/b2.html"),
		},
		"b1.html": &fstest.MapFile{
			Data: []byte("/c.html"),
		},
		"b2.html": &fstest.MapFile{
			Data: []byte("/c.html"),
		},
		"c.html": &fstest.MapFile{
			Data: []byte("/"),
		},
	}))
	defer srv.Close()
	cl := srv.Client()

	// Task fetches a page and extracts the URLs
	task := func(u string) ([]string, error) {
		res, err := cl.Get(srv.URL + u)
		if err != nil {
			return nil, err
		}
		defer res.Body.Close()
		if res.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("bad response: %q", u)
		}
		body, err := io.ReadAll(res.Body)
		if err != nil {
			return nil, err
		}

		return strings.Split(string(body), "\n"), nil
	}

	// results tracks which urls a url links to
	results := map[string][]string{}
	// queued tracks if a url has been queued to be fetched
	queued := map[string]bool{}

	// Manage the tasks with as many workers as GOMAXPROCS
	m := flowmatic.Manage(flowmatic.MaxProcs, task)
	m.Queue("/")
	// Work is like Exec but it halts if an error occurs
	for req, urls := range m.Work() {
		results[req] = urls
		for _, u := range urls {
			if !queued[u] {
				m.Queue(u)
				queued[u] = true
			}
		}
	}

	// Check for an error after processing!
	if m.HasErr() {
		fmt.Println("Error:", m.Error())
	}

	for _, key := range slices.Sorted(maps.Keys(results)) {
		fmt.Println(key, "links to:")
		for _, v := range results[key] {
			fmt.Println("- ", v)
		}
	}

	// Output:
	// / links to:
	// -  /a.html
	// /a.html links to:
	// -  /b1.html
	// -  /b2.html
	// /b1.html links to:
	// -  /c.html
	// /b2.html links to:
	// -  /c.html
	// /c.html links to:
	// -  /
}
