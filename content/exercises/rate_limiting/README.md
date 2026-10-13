# Rate limiting: token buckets

A *rate limiter* caps how many requests a client can make in a given time.
It protects your service from a noisy neighbour, a buggy retry loop or a
runaway script that fires thousands of requests a second. In this exercise you
build the classic algorithm, the token bucket, and put it in front of an HTTP
handler.

## Learn

### The token bucket

Picture a bucket that holds tokens:

- It holds at most `burst` tokens and starts full.
- Every request takes one token. No token, no request.
- Tokens drip back in at `rate` tokens per second, until the bucket is full
  again.

`burst` lets a client send a quick handful of requests. `rate` is the
long-run average it can keep up. Here is a bucket with `rate: 1` and
`burst: 3`:

| time | event | tokens before | allowed? | tokens after |
|---|---|---|---|---|
| 0s | request | 3 | yes | 2 |
| 0s | request | 2 | yes | 1 |
| 0s | request | 1 | yes | 0 |
| 0s | request | 0 | **no** | 0 |
| 0.5s | request | 0.5 | **no** | 0.5 |
| 1s | request | 1 | yes | 0 |
| 1h | request | 3 (full, not 3600) | yes | 2 |

### Refilling without a timer

You don't need a goroutine that adds a token every second. Remember when you
last refilled (`last`). On each request, work out how much time has passed
and add that many tokens in one go:

```go
now := b.now()
elapsed := now.Sub(b.last).Seconds() // a float64, such as 0.5
b.tokens = min(b.burst, b.tokens+elapsed*b.rate)
b.last = now
```

Go's built-in `min` returns the smaller of its arguments. Here it keeps the
bucket from overflowing past `burst`. Tokens are `float64`, so half a second
at `rate: 1` leaves half a token. That isn't enough for a request, but it
counts towards the next one.

The clock is a `now func() time.Time` field. Production passes `time.Now`;
the tests pass a fake clock that only moves when the test says so. Many
requests arrive at once, so the bucket guards its fields with a mutex, as in
*Mutexes*.

### 429 Too Many Requests

When a limiter says no, an HTTP server answers `429 Too Many Requests`. It
can also send a `Retry-After` header with a number of seconds, to tell
well-behaved clients how long to wait. Set headers *before* writing the
status:

```go
w.Header().Set("Retry-After", "1")
http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
```

`http.Error` writes the status code, then the message plus a newline, so the
body is `"rate limit exceeded\n"`.

### One bucket per client

A single shared bucket lets one noisy client use up the tokens for everybody.
Giving each client its own bucket, in a `map[string]*Bucket`, fixes that. A
server sees the client's address in `r.RemoteAddr` as `"host:port"`, such as
`"192.0.2.1:1234"`. The port changes with every new connection, so split it
off and key the map by the IP:

```go
ip, _, _ := net.SplitHostPort("192.0.2.1:1234") // ip = "192.0.2.1"
```

(Behind a load balancer, the real client IP usually arrives in a header such
as `X-Forwarded-For`. This exercise keeps to `RemoteAddr`.)

## Your task

The deploy API keeps falling over when a CI job gets stuck in a loop. Add rate
limiting in `exercise.go`.

### 1. Build the bucket

In `NewBucket`, start the bucket full: set `tokens` to the burst and `last` to
`now()`.

In `Allow`, which already locks the mutex:

1. Refill: add `elapsed seconds × rate` tokens, but never go above `burst`.
   Set `last` to the current time.
2. If there is at least one token, take one and return `true`. Otherwise
   return `false`.

```go
b := NewBucket(1, 2, clock.Now)
b.Allow() // => true
b.Allow() // => true
b.Allow() // => false
// clock moves forward 1s
b.Allow() // => true
```

### 2. Reject requests with 429

Complete `Middleware(b, next)`. Its tests use your bucket from task 1, so
finish that first. Return an `http.HandlerFunc` that calls `b.Allow()` for
every request:

- allowed: call `next.ServeHTTP(w, r)`;
- not allowed: don't call `next`. Set the `Retry-After` header to `"1"`, then
  respond with status `429` and the body `rate limit exceeded` using
  `http.Error`.

```text
POST /deploys  ->  200 deployed
POST /deploys  ->  200 deployed
POST /deploys  ->  429 rate limit exceeded      (burst 2, Retry-After: 1)
```

### 3. Give every client its own bucket

Complete `(*PerClient).Allow(r)`:

1. Get the client IP from `r.RemoteAddr` with `net.SplitHostPort`, and add
   `"net"` to the imports.
2. With `p.mu` locked, look the IP up in `p.buckets`. If it isn't there,
   create one with `NewBucket(p.rate, p.burst, p.now)` and store it.
3. Return that bucket's `Allow()`.

```text
burst 2:
192.0.2.1:1234     -> true
192.0.2.1:1234     -> true
192.0.2.1:1234     -> false
198.51.100.7:4000  -> true    (a different client)
192.0.2.1:5555     -> false   (same IP, different port: same client)
```

The tests also send requests from many goroutines at once, so `-race` will
catch a map used without the lock.

## Learn more

- [Wikipedia: Token bucket](https://en.wikipedia.org/wiki/Token_bucket)
- [MDN: 429 Too Many Requests](https://developer.mozilla.org/en-US/docs/Web/HTTP/Reference/Status/429)
- [MDN: Retry-After](https://developer.mozilla.org/en-US/docs/Web/HTTP/Reference/Headers/Retry-After)
- [pkg.go.dev: net.SplitHostPort](https://pkg.go.dev/net#SplitHostPort)
