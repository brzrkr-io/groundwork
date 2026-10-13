## Task 1: range over jobs

```go
for job := range jobs {
	results <- Result{Index: job.Index, Host: job.Host, Up: check(job.Host)}
}
```

## Task 2: n workers plus a closer

Start the `n` workers with a `sync.WaitGroup` instead of a bare `go`
(`for range n { wg.Go(...) }`). Then start one more goroutine that calls
`wg.Wait()` and then `close(results)`. Return `results`. Add `"sync"` to the
imports.

## Task 2: almost the answer

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

## Task 3: place by index

Make `out` with length `n` instead of capacity `n`:
`out := make([]Result, n)`. Then, inside the loop, replace the `append` with
`out[r.Index] = r`.
