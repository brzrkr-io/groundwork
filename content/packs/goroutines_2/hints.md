## Task 1: Fix the slow drain

Look at where `wg.Wait()` is. It blocks until every goroutine started so far has finished, so where you call it decides whether the drains overlap.

### Closer

`wg.Wait()` is inside the loop, so each drain must finish before the next one starts. Move it below the loop's closing `}`.

### Answer

```go
var wg sync.WaitGroup
for _, node := range nodes {
	wg.Go(func() {
		drain(node)
	})
}
wg.Wait()
```

Every drain starts first, then you wait once, so all of them run at the same time.

## Task 2: Ping hosts concurrently

This has the same shape as task 1, but each goroutine also has an answer to save. Every goroutine can write its answer into its own slot of `results`.

### Closer

Make a `sync.WaitGroup`, and inside `wg.Go(func() { ... })` write `results[i] = ping(host)`. Call `wg.Wait()` after the loop and before `return results`.

### Answer

```go
results := make([]bool, len(hosts))
var wg sync.WaitGroup
for i, host := range hosts {
	wg.Go(func() {
		results[i] = ping(host)
	})
}
wg.Wait()
return results
```

No two goroutines touch the same slot, and you only read `results` after `wg.Wait()`, so this is safe.
