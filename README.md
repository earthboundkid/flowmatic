# Workgroup [![GoDoc](https://pkg.go.dev/badge/github.com/carlmjohnson/workgroup)](https://pkg.go.dev/github.com/carlmjohnson/workgroup)
Workgroup is a generic Go concurrent task runner. It requires Go 1.20+.

## Execute heterogenous tasks
To execute heterogenous tasks with a set number of workers, use `workgroup.DoFuncs`:

```go
err := workgroup.DoFuncs(3,
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

## Execute homogenous tasks
To execute homogenous tasks with a set number of workers, use `workgroup.DoTasks`:

```go
things := []someType{thingA, thingB, thingC}

err := workgroup.DoTasks(len(things), things, func(thing someType) error {
    foo := thing.Frobincate()
    return foo.DoSomething()
})
```

## Manage tasks that spawn new tasks
For tasks that may create more work, use `workgroup.Do`.
Create a manager that will be serially executed,
and have it save the results
and examine the output of tasks to decide if there is more work to do.

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

// Manager keeps track of which pages have been visited and the results graph
manager := func(req string, links []string, err error) ([]string, error) {
    // Halt execution after the first error
    if err != nil {
        return nil, err
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
    return newpages, nil
}

// Process the tasks with as many workers as GOMAXPROCS
err := workgroup.Do(workgroup.MaxProcs, task, manager, "http://example.com/")
if err != nil {
    fmt.Println("error", err)
}
```
