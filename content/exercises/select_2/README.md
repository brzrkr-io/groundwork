# select: tickers and deadlines

Agents, exporters and controllers all run loops like "every 10 seconds, do
something, until told to stop". This exercise builds that loop with a ticker,
and fixes a classic timeout bug.

## Learn

### Recap

`select` waits on several channel operations and runs the first one that is
ready. `time.After(d)` returns a channel that receives once, after `d`.
`chan struct{}` is a signal channel; closing it wakes every receiver. The tests
use `testing/synctest`, so timers fire instantly in real time while the fake
clock still moves forward correctly.

### One timer, or a new timer every loop?

Each call to `time.After` creates a **new** timer. Inside a loop, that means
the timeout starts again from zero on every pass:

```go
for {
	select {
	case e := <-events:
		handle(e)
	case <-time.After(5 * time.Second): // a fresh 5s timer every time round
		return
	}
}
```

This only ends after 5 *quiet* seconds. If an event arrives every second, it
never ends. For an overall deadline, create the timer **once**, before the
loop:

```go
deadline := time.After(5 * time.Second) // one timer for the whole loop
for {
	select {
	case e := <-events:
		handle(e)
	case <-deadline:
		return
	}
}
```

### Tickers

A *ticker* sends the current time on its channel `C` again and again, every
interval. The first tick arrives after one interval, not straight away:

```go
ticker := time.NewTicker(10 * time.Second)
defer ticker.Stop() // release the ticker when the function returns

for {
	select {
	case <-ticker.C:
		scrapeMetrics()
	case <-stop:
		return
	}
}
```

```text
time:   0s ──── 10s ──── 20s ──── 25s
ticks:          tick     tick      │
stop:                              closed → return
```

`ticker.Stop()` stops future ticks. Using `defer` means you can't forget it,
whichever way the function returns.

## Your task

You're working on a small monitoring agent.

### 1. Fix the collection window

`CollectFor(events, window)` should collect events until `window` has passed
since it was called, or until `events` is closed, whichever comes first. It
returns the events in the order they arrived.

Right now, if an event arrives every second, `CollectFor` never stops early:
its timeout starts again after every event. Fix it so `window` is measured
from the start.

```go
// one event every second: "event 1", "event 2", ... "event 10", then closed
CollectFor(events, 3500*time.Millisecond)
// => []string{"event 1", "event 2", "event 3"}, returned after 3.5s

// "event 1", "event 2", then closed after 2s
CollectFor(events, 5*time.Second)
// => []string{"event 1", "event 2"}, returned after 2s
```

### 2. Send heartbeats

`Heartbeat(interval, stop, beat)` should call `beat()` once every `interval`
using a `time.Ticker`, until `stop` is closed. Then it stops the ticker and
returns how many times it called `beat`. The first beat happens after one
`interval`, not straight away.

```go
// stop is closed after 3.5s
Heartbeat(time.Second, stop, beat)
// beat is called at 1s, 2s and 3s; returns 3 at 3.5s
```

## Learn more

- [Go by Example: Tickers](https://gobyexample.com/tickers)
- [time.NewTicker](https://pkg.go.dev/time#NewTicker)
- [time.After](https://pkg.go.dev/time#After)
