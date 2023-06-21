# Workgroup [![GoDoc](https://pkg.go.dev/badge/github.com/carlmjohnson/workgroup)](https://pkg.go.dev/github.com/carlmjohnson/workgroup) [![Coverage Status](https://coveralls.io/repos/github/carlmjohnson/workgroup/badge.svg)](https://coveralls.io/github/carlmjohnson/workgroup) [![Go Report Card](https://goreportcard.com/badge/github.com/carlmjohnson/workgroup)](https://goreportcard.com/report/github.com/carlmjohnson/workgroup)

Workgroup is a generic Go library that provides a structured approach to concurrent programming. It lets you easily manage concurrent tasks in a manner that is predictable and scalable, and it provides a simple, yet effective approach to structuring concurrency.

Workgroup has a simple API consisting of three core functions: `Do`, `DoEach`, and `DoTasks`. It automatically handles spawning workers, collecting errors, and recovering from panics.

Workgroup requires Go 1.20+.

## How it works

### Execute heterogenous tasks
One problem that workgroup solves is managing the execution of multiple tasks in parallel that are independent of each other. For example, let's say you want to send data to three different downstream APIs. If any of the sends fail, you want to return an error. With traditional Go concurrency, this can quickly become complex and difficult to manage, with Goroutines, channels, and sync.WaitGroups to keep track of. Workgroup makes it simple.

To execute heterogenous tasks with a set number of workers, just use `workgroup.Do`:

<table>
<tr>
<th><code>workgroup</code></th>
<th><code>stdlib</code></th>
</tr>
<tr>
<td>

```go
err := workgroup.Do(3,
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
`workgroup.DoEach` is useful if you need to execute the same task on each item in a slice using a worker pool:

<table>
<tr>
<th><code>workgroup</code></th>
<th><code>stdlib</code></th>
</tr>
<tr>
<td>

```go
things := []someType{thingA, thingB, thingC}

err := workgroup.DoEach(numWorkers, things,
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
For tasks that may create more work, use `workgroup.DoTasks`.
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
workgroup.DoTasks(workgroup.MaxProcs, task, manager, "http://example.com/")
// Check if anything went wrong
if managerErr != nil {
    fmt.Println("error", managerErr)
}
```
