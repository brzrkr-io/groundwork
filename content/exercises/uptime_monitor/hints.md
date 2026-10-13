## Task 1: one loop, then the status

Loop over `results`. For a down result, `s.Down++` and `continue`. For an up
result, `s.Up++`, add its `Latency` to a running total, and remember its name
if `s.Slowest == ""` or its latency beats the slowest so far. After the loop,
divide the total by `time.Duration(s.Up)` when `s.Up > 0`. Pick `Status` with
a `switch`: `s.Down == 0` → `"OK"`, `s.Up == 0` → `"DOWN"`, otherwise
`"DEGRADED"`.

## Task 2: the request

```go
result := Result{Target: target, Attempts: 1}
ctx, cancel := context.WithTimeout(ctx, m.Timeout)
defer cancel()
req, err := http.NewRequestWithContext(ctx, http.MethodGet, target.URL, nil)
// handle err
start := time.Now()
resp, err := m.Client.Do(req)
result.Latency = time.Since(start)
// handle err, then resp.Body.Close() and check resp.StatusCode
```

## Task 3: a counted loop with a wait

`for attempt := 1; attempt <= 1+m.Retries; attempt++`. If `attempt > 1`,
`select` on `time.After(m.RetryDelay)` and `ctx.Done()` (return the previous
result in that case). Then `result = m.Check(ctx, target)`, set
`result.Attempts = attempt`, and `break` if `result.Up`.

## Task 4: indexes as jobs

```go
jobs := make(chan int)
go func() {
	defer close(jobs)
	for i := range targets {
		jobs <- i
	}
}()
results := make([]Result, len(targets))
var wg sync.WaitGroup
for range m.Workers {
	wg.Go(func() {
		for i := range jobs {
			results[i] = m.CheckWithRetry(ctx, targets[i])
		}
	})
}
wg.Wait()
return results
```

## Tasks 5 and 6: ticker loop, then one line

`Every`: make the ticker, `defer ticker.Stop()`, call `fn(ctx)` once, then
`for { select { case <-ticker.C: fn(ctx); case <-ctx.Done(): return ctx.Err() } }`.

`Run`:

```go
return Every(ctx, interval, func(ctx context.Context) {
	report(Summarize(m.CheckAll(ctx, targets)))
})
```
