package workgroup_test

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing/fstest"
	"time"

	"github.com/carlmjohnson/workgroup"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
)

func ExampleDo() {
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
	manager := func(req string, urls []string, err error) ([]string, error) {
		if err != nil {
			// If there's a problem fetching a page, try three times
			if tried[req] < 3 {
				tried[req]++
				return []string{req}, nil
			}
			return nil, err
		}
		results[req] = urls
		var newurls []string
		for _, u := range urls {
			if tried[u] == 0 {
				newurls = append(newurls, u)
				tried[u]++
			}
		}
		return newurls, nil
	}

	// Process the tasks with as many workers as runtime.NumGoroutine
	err := workgroup.Do(-1, task, manager, "/")
	if err != nil {
		fmt.Println("error", err)
	}

	keys := maps.Keys(results)
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

func ExampleDoTasks() {
	start := time.Now()
	task := func(d time.Duration) error {
		time.Sleep(d)
		fmt.Println("slept", d)
		return nil
	}
	err := workgroup.DoTasks(-1, task,
		50*time.Millisecond, 100*time.Millisecond, 200*time.Millisecond)
	if err != nil {
		fmt.Println("error", err)
	}
	fmt.Println("executed concurrently?", time.Since(start) < 300*time.Millisecond)
	// Output:
	// slept 50ms
	// slept 100ms
	// slept 200ms
	// executed concurrently? true
}

func ExampleDoFuncs() {
	start := time.Now()
	err := workgroup.DoFuncs(-1, func() error {
		time.Sleep(50 * time.Millisecond)
		fmt.Println("hello")
		return nil
	}, func() error {
		time.Sleep(100 * time.Millisecond)
		fmt.Println("world")
		return nil
	}, func() error {
		time.Sleep(200 * time.Millisecond)
		fmt.Println("from workgroup.DoFuncs")
		return nil
	})
	if err != nil {
		fmt.Println("error", err)
	}
	fmt.Println("executed concurrently?", time.Since(start) < 300*time.Millisecond)
	// Output:
	// hello
	// world
	// from workgroup.DoFuncs
	// executed concurrently? true
}
