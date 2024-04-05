package flowmatic_test

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing/fstest"

	"github.com/earthboundkid/flowmatic"
)

func ExampleAllTasks() {
	// Example site to crawl with recursive links
	srv := httptest.NewServer(http.FileServer(http.FS(fstest.MapFS{
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
	})))
	defer srv.Close()
	cl := srv.Client()

	// Task fetches a page and extracts the URLs
	task := func(u string) ([]string, error) {
		res, err := cl.Get(srv.URL + u)
		if err != nil {
			return nil, err
		}
		defer res.Body.Close()
		body, err := io.ReadAll(res.Body)
		if err != nil {
			return nil, err
		}

		return strings.Split(string(body), "\n"), nil
	}

	// Manager keeps track of which pages have been visited and the results graph
	tried := map[string]int{}
	results := map[string][]string{}

	// Process the tasks with as many workers as GOMAXPROCS
	it := flowmatic.Tasks(flowmatic.MaxProcs, task, "/")
	for r := range it {
		req := r.In
		urls := r.Out
		if r.HasErr() {
			// If there's a problem fetching a page, try three times
			if tried[req] < 3 {
				tried[req]++
				r.AddTask(req)
				continue
			}
			break
		}
		results[req] = urls
		for _, u := range urls {
			if tried[u] == 0 {
				r.AddTask(u)
				tried[u]++
			}
		}
	}

	keys := make([]string, 0, len(results))
	for key := range results {
		keys = append(keys, key)
	}
	slices.Sort(keys)
	for _, key := range keys {
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
