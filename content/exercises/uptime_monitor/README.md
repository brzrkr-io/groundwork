# Project: uptime monitor

Build the core of an uptime monitor, like a tiny Pingdom or Blackbox
Exporter. It checks a list of HTTP endpoints concurrently with a bounded
worker pool, gives every request a timeout, retries flaky endpoints, sums up
each round, and runs round after round until it is told to stop. This project
uses everything from this module.

## Learn

### How it fits together

```text
Run(ctx, targets, interval, report)
 └─ Every(ctx, interval, round)              schedule.go   task 5
     └─ round:
         CheckAll(ctx, targets)              pool.go       task 4
          ├─ worker ─┐
          ├─ worker ─┼─ CheckWithRetry       check.go      task 3
          └─ worker ─┘   └─ Check            check.go      task 2
                              └─ GET with per-attempt timeout
         Summarize(results)                  summary.go    task 1
         report(summary)
```

`ctx` flows from `Run` all the way down to each HTTP request. Cancelling it
stops the schedule, the waits between retries, and any request in flight.

### File map

| File | What's in it | You change |
|---|---|---|
| `monitor.go` | `Target`, `Result`, `Summary`, `Monitor` (settings) and `Run` | `Run` (task 6) |
| `summary.go` | `Summarize` | task 1 |
| `check.go` | `Check` and `CheckWithRetry` | tasks 2 and 3 |
| `pool.go` | `CheckAll` | task 4 |
| `schedule.go` | `Every` | task 5 |
| `uptime_test.go` | one test per task, in task order | nothing |

Read `monitor.go` first. Every type and field has a comment saying what it
means.

### Suggested order

Do the tasks in order. Each test only needs its own task and the ones before
it. `TestCheckAll` uses your `CheckWithRetry`, and `TestRun` uses everything.

### What you'll reuse

| From | Used for |
|---|---|
| *Context: cancelling HTTP requests* | `context.WithTimeout` + `http.NewRequestWithContext` |
| *Context and cancellation* | waiting with `select` on `time.After` and `ctx.Done()` |
| *Worker pools* | a `jobs` channel, `Workers` goroutines, results by index |
| *select: tickers and deadlines* | a `time.Ticker` loop with a stop case |

## Your task

### 1. Summarize a round

In `summary.go`, `Summarize(results)` fills in every field of `Summary`:

- `Up` and `Down` count the results with `Up` true and false.
- `AvgLatency` is the average `Latency` of the **up** results (total divided by
  `Up`), or 0 if none are up. Divide a `time.Duration` by a count with
  `total / time.Duration(n)`.
- `Slowest` is the `Target.Name` of the up result with the highest
  `Latency`. On a tie, the first one wins. It is `""` if none are up.
- `Status` is `"OK"` when nothing is down (including when there are no
  results), `"DOWN"` when something is down and nothing is up, and
  `"DEGRADED"` otherwise.

```go
Summarize([]Result{api up 10ms, billing down, search up 50ms})
// => Summary{Up: 2, Down: 1, AvgLatency: 30ms, Slowest: "search", Status: "DEGRADED"}
```

### 2. Check one target

In `check.go`, `m.Check(ctx, target)` makes **one** GET request to
`target.URL` with `m.Client`, and returns a `Result`:

- Create a context with `context.WithTimeout(ctx, m.Timeout)` for this
  attempt (and `defer cancel()`), and build the request with
  `http.NewRequestWithContext`.
- `Target` is `target`, and `Attempts` is `1`.
- `Latency` is how long `m.Client.Do` took: `time.Now()` before it,
  `time.Since(start)` after it.
- If `Do` returns an error (including a timeout), set `Err` to that error.
  `Status` stays 0 and `Up` stays false.
- Otherwise close the response body and set `Status`. A status from 200 to 299
  means `Up` is true and `Err` is nil. Any other status means `Up` is false
  and `Err` is `fmt.Errorf("unexpected status %d", status)`.

```go
m.Check(ctx, Target{Name: "api", URL: srv.URL + "/ok"})
// => Result{Target: ..., Up: true, Status: 200, Latency: 1.2ms, Attempts: 1, Err: nil}

m.Check(ctx, Target{Name: "billing", URL: srv.URL + "/down"})
// => Up false, Status 503, Err "unexpected status 503"

// /hang never answers; m.Timeout is 20ms
// => Up false, Status 0, errors.Is(Err, context.DeadlineExceeded)
```

### 3. Retry flaky targets

`m.CheckWithRetry(ctx, target)` calls `m.Check` up to `1 + m.Retries`
times, and stops as soon as an attempt is up. Before each retry (not before
the first attempt), wait `m.RetryDelay` using a `select` with
`time.After(m.RetryDelay)` and `ctx.Done()`. If `ctx` is cancelled while
waiting, stop and return the previous attempt's result. Return the last
attempt's `Result`, with `Attempts` set to the number of attempts made.

```go
// Retries: 2, and /flaky fails twice, then answers 200
m.CheckWithRetry(ctx, flaky) // => Up true, Attempts 3

// Retries: 2, and /down always fails
m.CheckWithRetry(ctx, down) // => Up false, Status 503, Attempts 3
```

### 4. Check everything with a worker pool

In `pool.go`, `m.CheckAll(ctx, targets)` checks every target with
`m.CheckWithRetry` and returns one `Result` per target, **in the same order
as `targets`**. Run exactly `m.Workers` worker goroutines that take jobs
from a channel, so at most `m.Workers` checks run at once.

A simple way to do it: send each target's **index** on an unbuffered `jobs`
channel from a goroutine (then close it). Each worker ranges over `jobs` and
writes `results[i]` directly. No two workers ever get the same `i`, so the
slice needs no lock. `wg.Wait()` before returning.

### 5. Run on a schedule

In `schedule.go`, `Every(ctx, interval, fn)` calls `fn(ctx)` once straight
away, then once on every tick of a `time.Ticker` with `interval`, until `ctx`
is cancelled. Then it stops the ticker and returns `ctx.Err()`.

```go
// interval 10s; ctx cancelled at 25s
Every(ctx, 10*time.Second, fn)
// fn is called at 0s, 10s and 20s; returns context.Canceled at 25s
```

### 6. Wire it together

In `monitor.go`, `m.Run(ctx, targets, interval, report)` uses `Every`. Each
round runs `m.CheckAll(ctx, targets)`, passes the results to `Summarize`,
and calls `report` with the summary. It returns what `Every` returns.

```go
m.Run(ctx, targets, 30*time.Second, func(s Summary) {
	fmt.Printf("%s: %d up, %d down, avg %v\n", s.Status, s.Up, s.Down, s.AvgLatency)
})
// DEGRADED: 2 up, 1 down, avg 12ms
// ...every 30s, until ctx is cancelled
```

## Learn more

- [Go blog: Go Concurrency Patterns: Context](https://go.dev/blog/context)
- [net/http: Client.Do](https://pkg.go.dev/net/http#Client.Do)
- [time.Ticker](https://pkg.go.dev/time#Ticker)
- [Go by Example: Tickers](https://gobyexample.com/tickers)
