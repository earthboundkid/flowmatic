package flowmatic_test

import (
	"crypto/md5"
	"fmt"
	"io"
	"io/fs"
	"maps"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
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
	m.Add("/")
	for range m.Exec() {
		req, urls, err := m.Result()
		if err != nil {
			// If there's a problem fetching a page, try three times
			if tried[req] < 3 {
				tried[req]++
				m.Add(req)
			}
			continue
		}
		results[req] = urls
		for _, u := range urls {
			if tried[u] == 0 {
				m.Add(u)
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
	// Read all the files in the file tree rooted at "testdata/md5all"
	// and returns a map from file path to the MD5 sum of the file's contents.
	// If the directory walk fails or any read operation fails,
	// print the error.
	//
	// Compare to https://pkg.go.dev/golang.org/x/sync/errgroup#example-Group-Pipeline and https://blog.golang.org/pipelines
	// This differs by doing all the filewalking before any digesting.
	manager := flowmatic.Manage(20, func(path string) (*[md5.Size]byte, error) {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		hash := md5.Sum(data)
		return &hash, nil
	})

	err := filepath.WalkDir("testdata/md5all", func(path string, info fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		manager.Add(path)

		return nil
	})
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	m := make(map[string]*[md5.Size]byte)
	maps.Insert(m, manager.Work())

	if manager.HasErr() {
		fmt.Println("Error:", manager.Err())
		return
	}

	for _, k := range slices.Sorted(maps.Keys(m)) {
		fmt.Printf("%s:\t%x\n", k, *m[k])
	}

	// Output:
	// testdata/md5all/hello.txt:	bea8252ff4e80f41719ea13cdf007273
}
