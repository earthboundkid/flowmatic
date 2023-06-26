# Flowmatic [![GoDoc](https://pkg.go.dev/badge/github.com/carlmjohnson/flowmatic)](https://pkg.go.dev/github.com/carlmjohnson/flowmatic) [![Coverage Status](https://coveralls.io/repos/github/carlmjohnson/flowmatic/badge.svg)](https://coveralls.io/github/carlmjohnson/flowmatic) [![Go Report Card](https://goreportcard.com/badge/github.com/carlmjohnson/flowmatic)](https://goreportcard.com/report/github.com/carlmjohnson/flowmatic)

![flowmatic](https://github.com/carlmjohnson/flowmatic/assets/222245/c14936e9-bb35-405b-926e-4cfeb8003439)


Flowmatic is a generic Go library that provides a [structured approach](https://vorpus.org/blog/notes-on-structured-concurrency-or-go-statement-considered-harmful/) to concurrent programming. It lets you easily manage concurrent tasks in a manner that is simple, yet effective and flexible.

Flowmatic has a easy to use API consisting of three core functions: `Do`, `DoEach`, and `DoTasks`. It automatically handles spawning workers, collecting errors, and propagating panics.

Flowmatic requires Go 1.20+.

## How it works

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


### Execute homogenous tasks
`flowmatic.DoEach` is useful if you need to execute the same task on each item in a slice using a worker pool:

<table>
<tr>
<th><code>flowmatic</code></th>
<th><code>stdlib</code></th>
</tr>
<tr>
<td>

```go
things := []someType{thingA, thingB, thingC}

err := flowmatic.DoEach(numWorkers, things,
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

### Manage tasks that spawn new tasks
For tasks that may create more work, use `flowmatic.DoTasks`.
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
flowmatic.DoTasks(flowmatic.MaxProcs, task, manager, "http://example.com/")
// Check if anything went wrong
if managerErr != nil {
    fmt.Println("error", managerErr)
}
```

Normally, it is very difficult to keep track of concurrent code because any combination of events could occur in any order or simultaneously, and each combination has to be accounted for by the programmer. `flowmatic.DoTasks` makes it simple to write concurrent code because everything follows a simple rule: **tasks happen concurrently; the manager runs serially**.

Centralizing control in the manager makes reasoning about the code radically simpler. When writing locking code, if you have M states and N methods, you need to think about all N states in each of the M methods, giving you an M × N code explosion. By centralizing the logic, the N states only need to be considered in one location: the manager.

## Note on panicking

In Go, if there is a panic in a goroutine, and that panic is not recovered, then the whole process is shutdown. There are pros and cons to this approach. The pro is that if the panic is the symptom of a programming error in the application, no further damage can be done by the application. The con is that in many cases, this leads to a shutdown in a situation that might be recoverable.

As a result, although the Go standard HTTP server will catch panics that occur in one of its HTTP handlers and continue serving requests, but the server cannot catch panics that occur in separate goroutines, and these will cause the whole server to go offline.

Flowmatic fixes this problem by catching a panic that occurs in one of its worker goroutines and repropagating it in the parent goroutine, so the panic can be caught and logged at the appropriate level.
