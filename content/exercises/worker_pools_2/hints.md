## Task 1: a closer goroutine

After starting the workers, start one more goroutine. It waits for the
WaitGroup and then closes `results`:

```go
go func() {
	wg.Wait()
	close(results)
}()
```

It has to be a goroutine. Calling `wg.Wait()` directly would block before
the collector loop starts, and the workers would be stuck sending.

## Task 2: a slice with one slot per image

Replace `var out []Result` with `out := make([]Result, len(images))`, and
replace the `append` with `out[p.index] = p.result`.
