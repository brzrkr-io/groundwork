# select: tickers and deadlines

Agents, exporters and controllers all run loops like "every 10 seconds, do
something, until told to stop". Here you build that loop with a *ticker*, and
fix a classic timeout bug.

**New here:** `time.NewTicker`, `ticker.C`, `ticker.Stop`, and creating a
timer once, outside a loop.

## Learn

### Recap

- `select` waits on several channel operations and runs the first one that
  is ready.
- `time.After(d)` returns a channel that receives one value after `d`.
- Closing a `chan struct{}` wakes every receiver. A closed channel stays
  ready forever.

The tests use `testing/synctest`: a fake clock, so a 10-second wait finishes
instantly but still counts as 10 seconds.

### One timer, or a new one every loop?

Every **call** to `time.After` starts a **new** timer. Put the call inside a
`select` in a loop, and the timer restarts from zero on every pass:

```go
for {
	select {
	case msg := <-alerts:
		page(msg)
	case <-time.After(30 * time.Second): // new 30s timer on every pass
		return
	}
}
```

That loop only ends after 30 *quiet* seconds. If an alert arrives every 10
seconds, it never ends.

For one overall deadline, call `time.After` **once**, before the loop. Store
the channel it returns in a variable, and receive from that variable in the
`case`. A channel is a value like any other, so you can keep it in a variable.

> **Why is it written like this?** Go requires `case <-ch:` to name a
> channel, but any expression that gives one works: a function call or a
> variable. Where you create the timer is *your* choice, and it decides
> what the timeout means.

### Tickers

A *ticker* sends on its channel `C` again and again, once per interval. The
first tick comes after one interval, not straight away.

```go
ticker := time.NewTicker(time.Minute)
defer ticker.Stop()

for {
	select {
	case <-ticker.C:
		rotateLogs()
	case <-quit:
		return
	}
}
```

```text
time:   0s ──── 1m ──── 2m ──── 2m30s
ticks:          tick    tick     │
quit:                            closed → return
```

> **Why is it written like this?** `time.NewTicker` and the field name `C`
> are fixed by the `time` package. Calling `defer ticker.Stop()` straight
> after creating it is convention: it releases the ticker however the
> function returns. The `for { select { … } }` shape is the standard Go loop
> for "wait for whichever happens first, forever".

## Your task

You're working on a small monitoring agent.

### 1. Fix the collection window

**Goal:** make `CollectFor(events, window)` stop once `window` has passed
since it was **called**, not since the last event.

It collects strings from `events`, in the order they arrive, and returns them
when either:

- `window` has passed since `CollectFor` started, or
- `events` is closed (`ok` is `false`),

whichever comes first. Right now its timeout restarts after every event, so a
steady stream of events keeps it running forever.

```go
// one event per second: "event 1", "event 2", ... "event 10", then closed
CollectFor(events, 3500*time.Millisecond)
// => []string{"event 1", "event 2", "event 3"}, returned after 3.5s

// "event 1", "event 2", then closed after 2s
CollectFor(events, 5*time.Second)
// => []string{"event 1", "event 2"}, returned after 2s
```

**Think it through:**
- Which line creates the timer, and how often does that line run?
- Where could it run exactly once, and how will the `case` reach its channel?

### 2. Send heartbeats

**Goal:** make `Heartbeat(interval, stop, beat)` call `beat()` once every
`interval` until `stop` is closed, then return how many times it called
`beat`.

- Use a `time.Ticker` with `interval`, and stop it with `defer`.
- On every tick: call `beat()` and add 1 to `beats`.
- When `stop` is closed: return `beats` straight away.
- The first beat comes after one `interval`, so stopping early can return 0.

```go
// stop is closed after 3.5s
Heartbeat(time.Second, stop, beat)
// beat is called at 1s, 2s and 3s; returns 3 at 3.5s

// stop is closed after 4s
Heartbeat(10*time.Second, stop, beat)
// returns 0 at 4s
```

**Think it through:**
- Which two things can happen next? Each one is a `case`.
- Compare with the ticker example in Learn: which names change?

## Learn more

- [Go by Example: Tickers](https://gobyexample.com/tickers)
- [time.NewTicker](https://pkg.go.dev/time#NewTicker)
- [time.After](https://pkg.go.dev/time#After)
