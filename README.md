# Flowmatic [![GoDoc](https://pkg.go.dev/badge/github.com/carlmjohnson/flowmatic)](https://pkg.go.dev/github.com/carlmjohnson/flowmatic) [![Coverage Status](https://coveralls.io/repos/github/carlmjohnson/flowmatic/badge.svg)](https://coveralls.io/github/carlmjohnson/flowmatic) [![Go Report Card](https://goreportcard.com/badge/github.com/carlmjohnson/flowmatic)](https://goreportcard.com/report/github.com/carlmjohnson/flowmatic) [![Mentioned in Awesome Go](https://awesome.re/mentioned-badge.svg)](https://github.com/avelino/awesome-go)

![Flowmatic logo](https://github.com/carlmjohnson/flowmatic/assets/222245/c14936e9-bb35-405b-926e-4cfeb8003439)


Flowmatic is a generic Go library that provides a [structured approach](https://vorpus.org/blog/notes-on-structured-concurrency-or-go-statement-considered-harmful/) to concurrent programming. It lets you easily manage concurrent tasks in a manner that is simple, yet effective and flexible.

Flowmatic has an easy to use API with functions for handling common concurrency patterns. It automatically handles spawning workers, collecting errors, and propagating panics.

Flowmatic requires Go 1.20+.

## Features

- Has a simple API that improves readability over channels/waitgroups/mutexes
- Handles a variety of concurrency problems such as heterogenous task groups, homogenous execution of a task over a slice, and dynamic work spawning
- Aggregates errors
- Properly propagates panics across goroutine boundaries
- Has helpers for context cancelation
- Few dependencies
- Good test coverage

## How to use Flowmatic

### Execute heterogenous tasks
One problem that Flowmatic solves is managing the execution of multiple tasks in parallel that are independent of each other. For example, let's say you want to send data to three different downstream APIs. If any of the sends fail, you want to return an error. With traditional Go concurrency, this can quickly become complex and difficult to manage, with Goroutines, channels, and `sync.WaitGroup`s to keep track of. Flowmatic makes it simple.

To execute heterogenous tasks, just use `flowmatic.Do`:

<table>
<tr>
<th><code>flowmatic</code></th>
<th><code>stdlib</code></th>
</tr>
<tr>
<td>

```go
err := flowmatic.Do(
    func() error {
        return doThingA(),
    },
    func() error {
        return doThingB(),
    },
    func() error {
        return doThingC(),
    })
```

</td>
<td>

```go
var wg sync.WaitGroup
var errs []error
errChan := make(chan error)

wg.Add(3)
go func() {
    defer wg.Done()
    if err := doThingA(); err != nil {
        errChan <- err
    }
}()
go func() {
    defer wg.Done()
    if err := doThingB(); err != nil {
        errChan <- err
    }
}()
go func() {
    defer wg.Done()
    if err := doThingC(); err != nil {
        errChan <- err
    }
}()

go func() {
    wg.Wait()
    close(errChan)
}()

for err := range errChan {
    errs = append(errs, err)
}

err := errors.Join(errs...)
```

</td>
</tr>
</table>

To create a context for tasks that is canceled on the first error,
use `flowmatic.All`.
To create a context for tasks that is canceled on the first success,
use `flowmatic.Race`.

```go
// Make variables to hold responses
var pageA, pageB, pageC string
// Race the requests to see who can answer first
err := flowmatic.Race(ctx,
	func(ctx context.Context) error {
		var err error
		pageA, err = request(ctx, "A")
		return err
	},
	func(ctx context.Context) error {
		var err error
		pageB, err = request(ctx, "B")
		return err
	},
	func(ctx context.Context) error {
		var err error
		pageC, err = request(ctx, "C")
		return err
	},
)
```

### Execute homogenous tasks
`flowmatic.Each` is useful if you need to execute the same task on each item in a slice using a worker pool:

<table>
<tr>
<th><code>flowmatic</code></th>
<th><code>stdlib</code></th>
</tr>
<tr>
<td>

```go
things := []someType{thingA, thingB, thingC}

err := flowmatic.Each(numWorkers, things,
    func(thing someType) error {
        foo := thing.Frobincate()
        return foo.DoSomething()
    })
```

</td>
<td>

```go
things := []someType{thingA, thingB, thingC}

work := make(chan someType)
errs := make(chan error)

for i := 0; i < numWorkers; i++ {
    go func() {
        for thing := range work {
            // Omitted: panic handling!
            foo := thing.Frobincate()
            errs <- foo.DoSomething()
        }
    }()
}

go func() {
    for _, thing := range things {
            work <- thing
    }

    close(tasks)
}()

var collectedErrs []error
for i := 0; i < len(things); i++ {
    collectedErrs = append(collectedErrs, <-errs)
}

err := errors.Join(collectedErrs...)
```

</td>
</tr>
</table>

Use `flowmatic.Map` to map an input slice to an output slice.

<table>
<tr>
<td colspan="2">

```go
func main() {
	results, err := Google(context.Background(), "golang")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	for _, result := range results {
		fmt.Println(result)
	}
}
```

</td></tr>
<tr>
<th><code>flowmatic</code></th>
<th><a href="https://pkg.go.dev/golang.org/x/sync/errgroup#example-Group-Parallel"><code>x/sync/errgroup</code></a></th>
</tr>
<tr><td>

```go
func Google(ctx context.Context, query string) ([]Result, error) {
	searches := []Search{Web, Image, Video}
	return flowmatic.Map(ctx, flowmatic.MaxProcs, searches,
		func(ctx context.Context, search Search) (Result, error) {
			return search(ctx, query)
		})
}
```


</td>
<td>

```go
func Google(ctx context.Context, query string) ([]Result, error) {
	g, ctx := errgroup.WithContext(ctx)

	searches := []Search{Web, Image, Video}
	results := make([]Result, len(searches))
	for i, search := range searches {
		i, search := i, search // https://golang.org/doc/faq#closures_and_goroutines
		g.Go(func() error {
			result, err := search(ctx, query)
			if err == nil {
				results[i] = result
			}
			return err
		})
	}
	if err := g.Wait(); err != nil {
		return nil, err
	}
	return results, nil
}
```

</td>
</table>

### Manage tasks that spawn new tasks
For tasks that may create more work, use `flowmatic.ManageTasks`.
Create a manager that will be serially executed,
and have it save the results
and examine the output of tasks to decide if there is more work to be done.

```go
// Task fetches a page and extracts the URLs
task := func(u string) ([]string, error) {
    page, err := getURL(ctx, u)
    if err != nil {
        return nil, err
    }
    return getLinks(page), nil
}

// Map from page to links
// Doesn't need a lock because only the manager touches it
results := map[string][]string{}
var managerErr error

// Manager keeps track of which pages have been visited and the results graph
manager := func(req string, links []string, err error) ([]string, bool) {
    // Halt execution after the first error
    if err != nil {
        managerErr = err
        return nil, false
    }
    // Save final results in map
    results[req] = urls

    // Check for new pages to scrape
    var newpages []string
    for _, link := range links {
        if _, ok := results[link]; ok {
            // Seen it, try the next link
            continue
        }
        // Add to list of new pages
        newpages = append(newpages, link)
        // Add placeholder to map to prevent double scraping
        results[link] = nil
    }
    return newpages, true
}

// Process the tasks with as many workers as GOMAXPROCS
flowmatic.ManageTasks(flowmatic.MaxProcs, task, manager, "http://example.com/")
// Check if anything went wrong
if managerErr != nil {
    fmt.Println("error", managerErr)
}
```

Normally, it is very difficult to keep track of concurrent code because any combination of events could occur in any order or simultaneously, and each combination has to be accounted for by the programmer. `flowmatic.ManageTasks` makes it simple to write concurrent code because everything follows a simple rule: **tasks happen concurrently; the manager runs serially**.

Centralizing control in the manager makes reasoning about the code radically simpler. When writing locking code, if you have M states and N methods, you need to think about all N states in each of the M methods, giving you an M × N code explosion. By centralizing the logic, the N states only need to be considered in one location: the manager.

### Advanced patterns with TaskPool

For very advanced uses, `flowmatic.TaskPool` takes the boilerplate out of managing a pool of workers. Compare Flowmatic to [this example from x/sync/errgroup](https://pkg.go.dev/golang.org/x/sync/errgroup#example-Group-Pipeline):

<table>
<tr>
<td colspan="2">

```go
func main() {
	m, err := MD5All(context.Background(), ".")
	if err != nil {
		log.Fatal(err)
	}

	for k, sum := range m {
		fmt.Printf("%s:\t%x\n", k, sum)
	}
}
```

</td></tr>
<tr>
<th><code>flowmatic</code></th>
<th><code>x/sync/errgroup</code></th>
</tr>
<tr><td>


```go
// MD5All reads all the files in the file tree rooted at root
// and returns a map from file path to the MD5 sum of the file's contents.
// If the directory walk fails or any read operation fails,
// MD5All returns an error.
func MD5All(ctx context.Context, root string) (map[string][md5.Size]byte, error) {
	// Make a pool of 20 digesters
	in, out := flowmatic.TaskPool(20, digest)

	m := make(map[string][md5.Size]byte)
	// Open two goroutines:
	// one for reading file names by walking the filesystem
	// one for recording results from the digesters in a map
	err := flowmatic.All(ctx,
		func(ctx context.Context) error {
			return walkFilesystem(ctx, root, in)
		},
		func(ctx context.Context) error {
			for r := range out {
				if r.Err != nil {
					return r.Err
				}
				m[r.In] = *r.Out
			}
			return nil
		},
	)

	return m, err
}

func walkFilesystem(ctx context.Context, root string, in chan<- string) error {
	defer close(in)

	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.Mode().IsRegular() {
			return nil
		}
		select {
		case in <- path:
		case <-ctx.Done():
			return ctx.Err()
		}

		return nil
	})
}

func digest(path string) (*[md5.Size]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	hash := md5.Sum(data)
	return &hash, nil
}
```

</td>
<td>

```go
type result struct {
	path string
	sum  [md5.Size]byte
}

// MD5All reads all the files in the file tree rooted at root and returns a map
// from file path to the MD5 sum of the file's contents. If the directory walk
// fails or any read operation fails, MD5All returns an error.
func MD5All(ctx context.Context, root string) (map[string][md5.Size]byte, error) {
	// ctx is canceled when g.Wait() returns. When this version of MD5All returns
	// - even in case of error! - we know that all of the goroutines have finished
	// and the memory they were using can be garbage-collected.
	g, ctx := errgroup.WithContext(ctx)
	paths := make(chan string)

	g.Go(func() error {
		defer close(paths)
		return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.Mode().IsRegular() {
				return nil
			}
			select {
			case paths <- path:
			case <-ctx.Done():
				return ctx.Err()
			}
			return nil
		})
	})

	// Start a fixed number of goroutines to read and digest files.
	c := make(chan result)
	const numDigesters = 20
	for i := 0; i < numDigesters; i++ {
		g.Go(func() error {
			for path := range paths {
				data, err := ioutil.ReadFile(path)
				if err != nil {
					return err
				}
				select {
				case c <- result{path, md5.Sum(data)}:
				case <-ctx.Done():
					return ctx.Err()
				}
			}
			return nil
		})
	}
	go func() {
		g.Wait()
		close(c)
	}()

	m := make(map[string][md5.Size]byte)
	for r := range c {
		m[r.path] = r.sum
	}
	// Check whether any of the goroutines failed. Since g is accumulating the
	// errors, we don't need to send them (or check for them) in the individual
	// results sent on the channel.
	if err := g.Wait(); err != nil {
		return nil, err
	}
	return m, nil
}
```

</td>
</tr>
</table>

## Note on panicking

In Go, if there is a panic in a goroutine, and that panic is not recovered, then the whole process is shutdown. There are pros and cons to this approach. The pro is that if the panic is the symptom of a programming error in the application, no further damage can be done by the application. The con is that in many cases, this leads to a shutdown in a situation that might be recoverable.

As a result, although the Go standard HTTP server will catch panics that occur in one of its HTTP handlers and continue serving requests, a standard Go HTTP server cannot catch panics that occur in separate goroutines, and these will cause the whole server to go offline.

Flowmatic fixes this problem by catching a panic that occurs in one of its worker goroutines and repropagating it in the parent goroutine, so the panic can be caught and logged at the appropriate level.
