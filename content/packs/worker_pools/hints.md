## Task 1: Write the worker

The worker keeps taking jobs until `jobs` is closed. Which loop over a
channel stops by itself when the channel is closed?

### Closer

Use `for job := range jobs { ... }` and send one `Result{Index: ..., Host:
..., Up: ...}` on `results` per job. Don't close `results`.

### Answer

```go
for job := range jobs {
	results <- Result{Index: job.Index, Host: job.Host, Up: check(job.Host)}
}
```

## Task 2: Close results when the workers are done

Someone has to close `results` after the last worker returns. A
`sync.WaitGroup` tracks the workers, and one extra goroutine can wait on it
so `startWorkers` still returns straight away.

### Closer

Start the workers with `wg.Go(func() { worker(jobs, results, check) })`
instead of `go`. Then `go func() { wg.Wait(); close(results) }()`. Add
`"sync"` to the imports.

### Answer

```go
results := make(chan Result)
var wg sync.WaitGroup
for range n {
	wg.Go(func() {
		worker(jobs, results, check)
	})
}
go func() {
	wg.Wait()
	close(results)
}()
return results
```

Calling `wg.Wait()` directly would deadlock, because nobody would be receiving yet.

## Task 3: Collect in order

`append` keeps arrival order. Each result already carries its `Index`, so
put it straight into the right place.

### Closer

Make `out` with length `n`, not just capacity: `make([]Result, n)`. In the
loop, assign to `out[r.Index]` instead of appending.

### Answer

```go
out := make([]Result, n)
for r := range results {
	out[r.Index] = r
}
return out
```
