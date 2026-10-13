## Task 1: where does Wait belong?

`wg.Wait()` blocks until every goroutine started so far has finished. Calling
it inside the loop means each drain must finish before the next one starts.
Start them all first, then wait once.

## Task 2: same shape as task 1

Make a `sync.WaitGroup`, start each ping with `wg.Go`, and write the answer
into `results[i]` inside the goroutine. Call `wg.Wait()` after the loop and
before `return results`.

## Task 2: almost the answer

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
