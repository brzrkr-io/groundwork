## Task 1: Summarize a round

One loop over `results` can count up and down targets, add up the latencies of
the up ones, and track the slowest. Work out the average and the `Status`
after the loop.

### Closer

For a down result, `s.Down++` and `continue`. For an up result, `s.Up++`, add
`r.Latency` to a `total`, and take its name when `s.Slowest == ""` or its
latency beats the slowest so far (`>` keeps the first one on a tie). Only
divide when `s.Up > 0`: `total / time.Duration(s.Up)`. A `switch` with no
value picks the status:

```go
switch {
case s.Down == 0:
	s.Status = "OK"
// ...
}
```

### Answer

```go
var s Summary
var total, slowest time.Duration
for _, r := range results {
	if !r.Up {
		s.Down++
		continue
	}
	s.Up++
	total += r.Latency
	if s.Slowest == "" || r.Latency > slowest {
		s.Slowest = r.Target.Name
		slowest = r.Latency
	}
}
if s.Up > 0 {
	s.AvgLatency = total / time.Duration(s.Up)
}
switch {
case s.Down == 0:
	s.Status = "OK"
case s.Up == 0:
	s.Status = "DOWN"
default:
	s.Status = "DEGRADED"
}
return s
```

Checking `s.Down == 0` first makes an empty round `"OK"`.

## Task 2: Check one target

Start with `Result{Target: target, Attempts: 1}`, give this attempt its own
timeout context, build the request with that context, and time only the
`m.Client.Do` call. An error from `Do` means no status; otherwise look at the
status code.

### Closer

```go
ctx, cancel := context.WithTimeout(ctx, m.Timeout)
defer cancel()
req, err := http.NewRequestWithContext(ctx, http.MethodGet, target.URL, nil)
start := time.Now()
resp, err := m.Client.Do(req)
result.Latency = time.Since(start)
```

After a successful `Do`, call `resp.Body.Close()` and set `result.Status`.
A status outside 200–299 gets `fmt.Errorf("unexpected status %d", ...)`.
Add `"fmt"`, `"net/http"` and `"time"` to the imports.

### Answer

```go
result := Result{Target: target, Attempts: 1}
ctx, cancel := context.WithTimeout(ctx, m.Timeout)
defer cancel()
req, err := http.NewRequestWithContext(ctx, http.MethodGet, target.URL, nil)
if err != nil {
	result.Err = err
	return result
}
start := time.Now()
resp, err := m.Client.Do(req)
result.Latency = time.Since(start)
if err != nil {
	result.Err = err
	return result
}
resp.Body.Close()
result.Status = resp.StatusCode
if resp.StatusCode < 200 || resp.StatusCode > 299 {
	result.Err = fmt.Errorf("unexpected status %d", resp.StatusCode)
	return result
}
result.Up = true
return result
```

## Task 3: Retry flaky targets

Use a counted loop from attempt 1 to `1 + m.Retries`. Wait only before a
retry, never before the first attempt, and stop as soon as one attempt is up.

### Closer

Inside the loop, when `attempt > 1`, wait with a `select`:

```go
select {
case <-time.After(m.RetryDelay):
case <-ctx.Done():
	return result // the previous attempt
}
```

Then `result = m.Check(ctx, target)`, set `result.Attempts = attempt`, and
`break` if `result.Up`.

### Answer

```go
var result Result
for attempt := 1; attempt <= 1+m.Retries; attempt++ {
	if attempt > 1 {
		select {
		case <-time.After(m.RetryDelay):
		case <-ctx.Done():
			return result
		}
	}
	result = m.Check(ctx, target)
	result.Attempts = attempt
	if result.Up {
		break
	}
}
return result
```

## Task 4: Check everything with a worker pool

Send each target's index on a `jobs` channel from one goroutine, then close
it. Start exactly `m.Workers` goroutines that range over `jobs` and write
into a results slice at that index, then wait for them all.

### Closer

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
		// range over jobs and fill results[i]
	})
}
```

Add `"sync"` to the imports.

### Answer

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

Each index goes to only one worker, so no two goroutines write the same slot
and the slice needs no lock.

## Task 5: Run on a schedule

Make a `time.Ticker`, make sure it is stopped when `Every` returns, and call
`fn` once before the loop. Then loop forever, waiting for either a tick or
cancellation.

### Closer

```go
ticker := time.NewTicker(interval)
defer ticker.Stop()
fn(ctx)
for {
	select {
	case <-ticker.C:
		// ...
	case <-ctx.Done():
		// ...
	}
}
```

### Answer

```go
ticker := time.NewTicker(interval)
defer ticker.Stop()
fn(ctx)
for {
	select {
	case <-ticker.C:
		fn(ctx)
	case <-ctx.Done():
		return ctx.Err()
	}
}
```

## Task 6: Wire it together

`Run` is one call to `Every`. The function you pass to it does a full round:
check all targets, summarize, report.

### Closer

```go
return Every(ctx, interval, func(ctx context.Context) {
	// CheckAll, then Summarize, then report
})
```

### Answer

```go
return Every(ctx, interval, func(ctx context.Context) {
	report(Summarize(m.CheckAll(ctx, targets)))
})
```
