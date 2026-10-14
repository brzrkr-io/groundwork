## Task 1: Run everything and report every error

A goroutine can't return its error, so give each task its own slot in a
slice and fill it by index. After everyone is done, `errors.Join` turns the
slice into one error.

### Closer

`errs := make([]error, len(tasks))` and a `sync.WaitGroup`. In
`for i, task := range tasks`, `wg.Go(func() { errs[i] = task() })`. After
`wg.Wait()`, `return errors.Join(errs...)`. Add `"errors"` and `"sync"` to
the imports.

### Answer

```go
errs := make([]error, len(tasks))
var wg sync.WaitGroup
for i, task := range tasks {
	wg.Go(func() {
		errs[i] = task()
	})
}
wg.Wait()
return errors.Join(errs...)
```

`errors.Join` skips `nil`s and returns `nil` when every task succeeded.

## Task 2: Run at most limit at once

Start from your task 1 code. Add a buffered channel with `limit` slots: take
a slot before starting each task, and give it back when the task finishes.

### Closer

`sem := make(chan struct{}, limit)`. In the loop, before `wg.Go`, send
`sem <- struct{}{}`. Inside the goroutine, start with
`defer func() { <-sem }()`.

### Answer

```go
errs := make([]error, len(tasks))
sem := make(chan struct{}, limit)
var wg sync.WaitGroup
for i, task := range tasks {
	sem <- struct{}{} // wait for a free slot
	wg.Go(func() {
		defer func() { <-sem }()
		errs[i] = task()
	})
}
wg.Wait()
return errors.Join(errs...)
```

The send blocks while `limit` tasks hold slots, so no more than `limit` run at once.

## Task 3: Check hosts

Don't write the concurrency again. Build a `func() error` task for each host
that calls `check` and adds the host name to any error, then hand the tasks to
`RunLimited`.

### Closer

In `for _, host := range hosts`, append a closure that calls `check(host)`.
On error, return `fmt.Errorf("%s: %w", host, err)`; otherwise return `nil`.
Finish with `return RunLimited(tasks, limit)`. Add `"fmt"` to the imports.

### Answer

```go
var tasks []func() error
for _, host := range hosts {
	tasks = append(tasks, func() error {
		if err := check(host); err != nil {
			return fmt.Errorf("%s: %w", host, err)
		}
		return nil
	})
}
return RunLimited(tasks, limit)
```

Returning plain `nil` for healthy hosts lets `errors.Join` skip them, and `%w` keeps `errors.Is` working.
