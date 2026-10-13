# Project: uptime monitor

Build the core of an uptime monitor, like a tiny Pingdom or Blackbox
Exporter. It checks HTTP endpoints with a bounded worker pool, gives each
request a timeout, retries flaky endpoints, sums up each round, and repeats
until told to stop. This project uses everything from this module.

**New here:** nothing; practice. You combine pieces from earlier exercises.

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

`ctx` flows from `Run` down to each HTTP request. Cancelling it stops the
schedule, the waits between retries, and any request in flight.

> **Why is it written like this?** Passing `ctx` as the **first** parameter
> is a Go convention every library follows. Splitting the code into five files
> is the author's choice: Go only requires them to be in the same package.

### File map

| File | What's in it | You change |
|---|---|---|
| `monitor.go` | `Target`, `Result`, `Summary`, `Monitor` (settings) and `Run` | `Run` (task 6) |
| `summary.go` | `Summarize` | task 1 |
| `check.go` | `Check` and `CheckWithRetry` | tasks 2 and 3 |
| `pool.go` | `CheckAll` | task 4 |
| `schedule.go` | `Every` | task 5 |
| `uptime_test.go` | one test per task, in task order | nothing |

Read `monitor.go` first: every type and field has a comment. Do the tasks in
order; each test needs only its task and earlier ones (`TestCheckAll` uses
your `CheckWithRetry`; `TestRun` uses everything).

### What you'll reuse

| From | Used for |
|---|---|
| *Context: cancelling HTTP requests* | `context.WithTimeout` + `http.NewRequestWithContext` |
| *Context and cancellation* | waiting with `select` on `time.After` and `ctx.Done()` |
| *Worker pools* | a `jobs` channel, `Workers` goroutines, results by index |
| *select: tickers and deadlines* | a `time.Ticker` loop with a stop case |

### Two reminders

Dividing a `time.Duration` by a count needs both sides to be `Duration`:

```go
perLap := raceTime / time.Duration(laps)
```

Each starter file imports only what it uses now. When you add code that
needs `fmt`, `net/http`, `time` or `sync`, add that import (Go rejects both
missing and unused imports).

## Your task

Fill in the six TODOs so the monitor checks targets every `interval` and
reports a summary of each round.

### 1. Summarize a round

**Goal:** in `summary.go`, make `Summarize(results)` fill in every field of
`Summary`.

- `Up` / `Down`: how many results have `Up` true / false.
- `AvgLatency`: total `Latency` of the **up** results divided by `Up`; 0 if
  none are up.
- `Slowest`: `Target.Name` of the up result with the highest `Latency`. On
  a tie the first one wins. `""` if none are up.
- `Status`:

| Situation | Status |
|---|---|
| nothing down (also: no results at all) | `"OK"` |
| something down, nothing up | `"DOWN"` |
| some up, some down | `"DEGRADED"` |

```go
Summarize([]Result{api up 10ms, billing down, search up 50ms})
// => Summary{Up: 2, Down: 1, AvgLatency: 30ms, Slowest: "search", Status: "DEGRADED"}
```

**Think it through:**
- Which totals do you need to keep while looping, besides the `Summary` itself?
- Down results must not affect `AvgLatency` or `Slowest`: how do you skip them?
- "First one wins on a tie": `>` or `>=`?

### 2. Check one target

**Goal:** in `check.go`, make `m.Check(ctx, target)` do **one** GET of
`target.URL` with `m.Client` and return a `Result`.

- Give this attempt its own timeout: a context from
  `context.WithTimeout(ctx, m.Timeout)` (and `defer cancel()`), and a request
  built with `http.NewRequestWithContext`.
- `Target` is `target`; `Attempts` is `1`.
- `Latency` is how long `m.Client.Do` took: `time.Now()` before it,
  `time.Since(start)` after it.
- `Do` returns an error (timeouts too) → `Err` is that error; `Status` stays
  0, `Up` stays false.
- Otherwise close the response body and set `Status`. 200–299 → `Up` true,
  `Err` nil. Any other status → `Up` false, `Err` is
  `fmt.Errorf("unexpected status %d", status)`.

```go
m.Check(ctx, Target{Name: "api", URL: srv.URL + "/ok"})
// => Result{Target: ..., Up: true, Status: 200, Latency: 1.2ms, Attempts: 1, Err: nil}

m.Check(ctx, Target{Name: "billing", URL: srv.URL + "/down"})
// => Up false, Status 503, Err "unexpected status 503"

// /hang never answers; m.Timeout is 20ms
// => Up false, Status 0, errors.Is(Err, context.DeadlineExceeded)
```

**Think it through:**
- Which fields of `Result` can you set before making the request?
- Where does the status code live on the response, and which range means up?

### 3. Retry flaky targets

**Goal:** make `m.CheckWithRetry(ctx, target)` call `m.Check` up to
`1 + m.Retries` times, stopping as soon as an attempt is up.

- Before each **retry** (not the first attempt), wait `m.RetryDelay` with a
  `select` on `time.After(m.RetryDelay)` and `ctx.Done()`.
- If `ctx` is cancelled while waiting, stop and return the previous
  attempt's result.
- Return the last attempt's `Result`, with `Attempts` set to how many
  attempts were made.

```go
// Retries: 2, and /flaky fails twice, then answers 200
m.CheckWithRetry(ctx, flaky) // => Up true, Attempts 3

// Retries: 2, and /down always fails
m.CheckWithRetry(ctx, down) // => Up false, Status 503, Attempts 3
```

**Think it through:**
- The result must outlive one loop pass: where do you declare it?
- `Check` always sets `Attempts` to 1. Which number should overwrite it?

### 4. Check everything with a worker pool

**Goal:** in `pool.go`, make `m.CheckAll(ctx, targets)` return one
`m.CheckWithRetry` result per target, **in the same order as `targets`**.

Run exactly `m.Workers` worker goroutines taking jobs from a channel, so at
most `m.Workers` checks run at once. One simple design:

- A goroutine sends each target's **index** on an unbuffered `jobs`
  channel, then closes it.
- Each worker ranges over `jobs` and writes `results[i]` directly. No two
  workers get the same `i`, so the slice needs no lock.
- `wg.Wait()` before returning.

**Think it through:**
- How long must `results` be, and when do you create it?
- Why does sending indexes (not targets) keep the order right?

### 5. Run on a schedule

**Goal:** in `schedule.go`, make `Every(ctx, interval, fn)` call `fn(ctx)`
right away, then on every tick, until `ctx` is cancelled.

Use a `time.Ticker` with `interval`. When `ctx` is cancelled, stop the
ticker and return `ctx.Err()`.

```go
// interval 10s; ctx cancelled at 25s
Every(ctx, 10*time.Second, fn)
// fn is called at 0s, 10s and 20s; returns context.Canceled at 25s
```

**Think it through:**
- The first call happens before any tick: is it inside or before the loop?
- Which two channels does the `select` wait on?

### 6. Wire it together

**Goal:** in `monitor.go`, make `m.Run(ctx, targets, interval, report)` use
`Every` to run rounds, and return what `Every` returns.

Each round: `m.CheckAll(ctx, targets)` → pass the results to `Summarize` →
call `report` with the summary.

```go
m.Run(ctx, targets, 30*time.Second, func(s Summary) {
	fmt.Printf("%s: %d up, %d down, avg %v\n", s.Status, s.Up, s.Down, s.AvgLatency)
})
// DEGRADED: 2 up, 1 down, avg 12ms
// ...every 30s, until ctx is cancelled
```

**Think it through:**
- `Every` wants a function. Look at the type of its `fn` parameter in
  `schedule.go`: what must your round function accept?
- Each step's output is the next step's input.

## Learn more

- [Go blog: Go Concurrency Patterns: Context](https://go.dev/blog/context)
- [net/http: Client.Do](https://pkg.go.dev/net/http#Client.Do)
- [time.Ticker](https://pkg.go.dev/time#Ticker)
- [Go by Example: Tickers](https://gobyexample.com/tickers)
