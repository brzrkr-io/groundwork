# Rate limiting: fix the bucket

Your team's public webhook receiver has a per-client rate limiter, but it lets
far too much through, its `Retry-After` header is always `1`, and its map of
clients grows forever as internet scanners come and go. Time to fix all three.

## Learn

### Recap

As in *Rate limiting: token buckets*, a bucket holds at most `burst` tokens,
and each request takes one. On every call it refills itself for the time
since `last`, at `rate` tokens per second, using `min` so it never goes above
`burst`. A `Limiter` keeps one bucket per client IP in a map guarded by a
mutex. The clock is an injected `now func() time.Time`.

### How long until the next token?

`Retry-After` should tell a client how long to wait, not just say "1". If a
bucket has `tokens` (less than 1), it is missing `1 - tokens` of a token, and
it earns `rate` tokens per second:

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

To turn seconds (a `float64`) into a `time.Duration`, multiply by
`float64(time.Second)`, which is the number of nanoseconds in a second:

```go
time.Duration(1.5 * float64(time.Second)) // 1.5s
```

`Retry-After` only takes whole seconds. Round *up*, so that clients never
come back too early. `math.Ceil` rounds a `float64` up, and `strconv.Itoa`
turns an `int` into text:

```go
secs := math.Ceil(d.Seconds()) // 1.5s -> 2.0
strconv.Itoa(int(secs))        // "2"
```

### Forgetting idle clients

Every new IP adds an entry to the map. A scanner that tries 100,000 addresses
leaves 100,000 buckets behind, and memory grows until the pod is OOM-killed.
The fix is to delete clients you haven't seen for a while. Each `client`
records `seen`, the last time it made a request.

Deleting from a map while ranging over it is allowed in Go:

```go
for ip, c := range l.clients {
	if tooOld(c) {
		delete(l.clients, ip)
	}
}
```

A real service would call `Cleanup` every minute from a goroutine with a
`time.Ticker`, while requests are being served, so it must hold `l.mu`.

## Your task

### 1. Fix the refill

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

### 2. Report an accurate Retry-After

Complete `RetryAfter`. Lock `b.mu`, call `b.refill()`, then return `0` if
there is at least one whole token. Otherwise return `(1 - tokens) / rate`
seconds as a `time.Duration`. `rate` is always above 0.

Then, in `Middleware`, replace the fixed `"1"` with `b.RetryAfter()`, rounded
up to whole seconds with `math.Ceil`. Add `"math"` and `"strconv"` to the
imports.

```text
rate 0.5, burst 1:
request at 0s    -> 200
request at 0s    -> 429, Retry-After: 2    (2s to go)
request at 0.5s  -> 429, Retry-After: 2    (1.5s to go)
request at 1.5s  -> 429, Retry-After: 1    (0.5s to go)
request at 2s    -> 200
```

### 3. Evict idle clients

Complete `Cleanup(idle)`. With `l.mu` locked, delete every client whose
`seen` time is `idle` or more before `l.now()`, and return how many you
deleted.

```text
9:00:00  requests from 192.0.2.1 and 192.0.2.2
9:00:30  requests from 192.0.2.1 and 192.0.2.3
9:01:00  Cleanup(1m) => 1   (192.0.2.2 was idle for exactly 1m)
9:01:30  Cleanup(1m) => 2   (192.0.2.1 and 192.0.2.3)
```

## Learn more

- [MDN: Retry-After](https://developer.mozilla.org/en-US/docs/Web/HTTP/Reference/Headers/Retry-After)
- [pkg.go.dev: math.Ceil](https://pkg.go.dev/math#Ceil)
- [Go spec: For statements with range clause](https://go.dev/ref/spec#For_range)
