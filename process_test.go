package workgroup_test

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing/fstest"

	"github.com/carlmjohnson/workgroup"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
)

func ExampleProcess() {
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

	// Manager keeps track of which pages have been visited and results graph
	seen := map[string]bool{}
	results := map[string][]string{}

	err := workgroup.Process(-1, func(u string) ([]string, error) {
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
	}, func(req string, urls []string, err error) ([]string, error) {
		if err != nil {
			return nil, err
		}
		results[req] = urls
		var newurls []string
		for _, u := range urls {
			if !seen[u] {
				newurls = append(newurls, u)
				seen[u] = true
			}
		}
		return newurls, nil
	}, "/")

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
