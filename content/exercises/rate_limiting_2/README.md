# Rate limiting: fix the bucket

Your team's public webhook receiver has a per-client rate limiter, but it
lets far too much through, its `Retry-After` is always `1`, and its map of
clients grows forever as internet scanners come and go. Fix all three.

**New here:** `float64` seconds to `time.Duration`, `math.Ceil`,
`strconv.Itoa`, `delete` while ranging over a map.

## Learn

### Recap

As in *Rate limiting: token buckets*, a bucket holds at most `burst` tokens
and each request takes one. On every call it refills for the time since
`last`, at `rate` tokens per second, and never goes above `burst`. A
`Limiter` keeps one bucket per client IP in a map guarded by a mutex. The
clock is an injected `now func() time.Time`.

### How long until the next token?

If a bucket has `tokens` (less than 1), it is missing `1 - tokens` of a
token, and it earns `rate` tokens per second:

```text
wait (seconds) = (1 - tokens) / rate
```

| rate | tokens | wait |
|---|---|---|
| 1 | 0 | 1s |
| 1 | 0.25 | 750ms |
| 0.5 | 0 | 2s |
| 0.5 | 0.25 | 1.5s |
| 2 | 0 | 500ms |

A `time.Duration` is a count of nanoseconds. To turn seconds held in a
`float64` into one, multiply by `float64(time.Second)` and convert:

```go
time.Duration(0.25 * float64(time.Minute)) // 15s
```

> **Why is it written like this?** Go requires the conversions: it never
> mixes `float64` and `time.Duration` in one expression by itself.

### Whole seconds, rounded up

`Retry-After` takes whole seconds. Round **up**, so clients never come back
too early. `math.Ceil` rounds a `float64` up, and `strconv.Itoa` turns an
`int` into text:

```go
math.Ceil(2.1)       // 3.0 (still a float64)
strconv.Itoa(int(x)) // "3" when x is 3.0
d.Seconds()          // a Duration as float64 seconds
```

### Forgetting idle clients

A scanner that tries 100,000 addresses leaves 100,000 buckets behind, and
memory grows until the pod is OOM-killed. So delete clients you haven't seen
for a while. Each `client` records `seen`, the time of its last request.

Go allows deleting from a map while ranging over it:

```go
for name, size := range files {
	if size == 0 {
		delete(files, name)
	}
}
```

A real service calls `Cleanup` every minute from a goroutine while requests
are served, so it must hold `l.mu`.

## Your task

### 1. Fix the refill

**Goal:** make `refill` stop at `burst` and stop counting the same time
twice.

`refill` has two bugs:

- it can fill the bucket past `burst`, so an idle client gets a huge burst;
- it never updates `last`, so the same stretch of time is added again on
  every call.

```go
b := NewBucket(1, 1, clock.Now)
b.Allow() // => true
// clock moves forward 1s
b.Allow() // => true
b.Allow() // => false (the buggy version says true)
```

**Think it through:**
- Which built-in function keeps a number from going above a limit?
- `refill` already has `now`. Which field should hold it afterwards?

### 2. Report an accurate Retry-After

**Goal:** make `RetryAfter` return how long until the next token, and send
it from `Middleware` in whole seconds.

In `RetryAfter`: lock `b.mu`, call `b.refill()`, then return `0` if there is
at least one whole token. Otherwise return `(1 - tokens) / rate` seconds as
a `time.Duration`. `rate` is always above 0.

In `Middleware`, replace the fixed `"1"` with `b.RetryAfter()` rounded up to
whole seconds. Add `"math"` and `"strconv"` to the imports.

```text
rate 0.5, burst 1:
request at 0s    -> 200
request at 0s    -> 429, Retry-After: 2    (2s to go)
request at 0.5s  -> 429, Retry-After: 2    (1.5s to go)
request at 1.5s  -> 429, Retry-After: 1    (0.5s to go)
request at 2s    -> 200
```

**Think it through:**
- How do you make sure the lock is released on both `return`s?
- In `Middleware` the chain is: Duration → float64 seconds → rounded up →
  `int` → text. Which function does each step?

### 3. Evict idle clients

**Goal:** make `Cleanup(idle)` delete idle clients and return how many it
deleted.

With `l.mu` locked, delete every client whose `seen` time is `idle` **or
more** before `l.now()`.

```text
9:00:00  requests from 192.0.2.1 and 192.0.2.2
9:00:30  requests from 192.0.2.1 and 192.0.2.3
9:01:00  Cleanup(1m) => 1   (192.0.2.2 was idle for exactly 1m)
9:01:30  Cleanup(1m) => 2   (192.0.2.1 and 192.0.2.3)
```

**Think it through:**
- "Exactly 1m" counts as idle. Which comparison operator does that need?
- You need a counter that starts at 0 before the loop.

## Learn more

- [MDN: Retry-After](https://developer.mozilla.org/en-US/docs/Web/HTTP/Reference/Headers/Retry-After)
- [pkg.go.dev: math.Ceil](https://pkg.go.dev/math#Ceil)
- [Go spec: For statements with range clause](https://go.dev/ref/spec#For_range)
