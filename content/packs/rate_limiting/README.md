# Rate limiting: token buckets

A *rate limiter* caps how many requests a client can make in a given time.
It protects your service from a noisy neighbour, a buggy retry loop or a
runaway script. You build the classic algorithm, the token bucket, and put
it in front of an HTTP handler.

**New here:** the token bucket, `429` with `Retry-After`. (`min`,
`http.HandlerFunc` middleware and `net.SplitHostPort` are practice.)

## Learn

### The token bucket

Picture a bucket that holds tokens:

- It holds at most `burst` tokens and starts full.
- Every request takes one token. No token, no request.
- Tokens drip back in at `rate` tokens per second, until it is full again.

`burst` is a quick handful of requests; `rate` is the long-run average. A
bucket with `rate: 1` and `burst: 3`:

| time | tokens before | allowed? | tokens after |
|---|---|---|---|
| 0s | 3 | yes | 2 |
| 0s | 2 | yes | 1 |
| 0s | 1 | yes | 0 |
| 0s | 0 | **no** | 0 |
| 0.5s | 0.5 | **no** | 0.5 |
| 1s | 1 | yes | 0 |
| 1h | 3 (full, not 3600) | yes | 2 |

### Refilling without a timer

You don't need a goroutine that adds a token every second. Remember when you
last refilled (`last`). On each request, work out how many seconds have
passed, multiply by `rate`, and add that many tokens in one go. Then move
`last` to now, so the same time isn't counted twice.

Two tools help:

```go
gap := later.Sub(earlier).Seconds() // a float64, such as 0.5
smaller := min(10.0, 3.5)           // 3.5: the built-in min
```

Tokens are `float64`, so half a second at `rate: 1` leaves half a token. It
isn't enough for a request, but it counts towards the next one.

> **Why is it written like this?** `min` is built into Go (1.21+), no import
> needed; its arguments must be the same type. Storing tokens as `float64` is
> the author's choice, so partial tokens add up.

The clock is a `now func() time.Time` field. Production passes `time.Now`;
the tests pass a fake clock that only moves when the test says so. The
bucket guards its fields with a mutex, as in *Mutexes*.

### Middleware and 429 Too Many Requests

*Middleware* wraps a handler: it runs first, then decides whether to call
the wrapped handler. `http.HandlerFunc(f)` turns a plain function into an
`http.Handler`:

```go
func logged(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.URL.Path)
		next.ServeHTTP(w, r)
	})
}
```

When a limiter says no, the server answers `429 Too Many Requests`, often
with a `Retry-After` header holding a number of seconds. Set headers
**before** writing the status. `http.Error(w, msg, code)` writes the status,
then `msg` plus a newline:

```go
w.Header().Set("Cache-Control", "no-store")
http.Error(w, "not found", http.StatusNotFound) // body: "not found\n"
```

> **Why is it written like this?** Go requires the headers first: once the
> status is written, later `Header().Set` calls are ignored. Named constants
> like `http.StatusTooManyRequests` instead of `429` are convention.

### One bucket per client

One shared bucket lets a noisy client use up everybody's tokens. Give each
client its own bucket in a `map[string]*Bucket`. `r.RemoteAddr` is
`"host:port"`, and the port changes on every connection, so key the map by
the host only:

```go
host, port, err := net.SplitHostPort("example.com:443")
// host = "example.com", port = "443"
```

Use `_` for results you don't need. (Behind a load balancer the real IP
arrives in a header such as `X-Forwarded-For`; this exercise keeps to
`RemoteAddr`.)

## Your task

The deploy API keeps falling over when a CI job gets stuck in a loop. Add
rate limiting in `exercise.go`.

### 1. Build the bucket

**Goal:** make `NewBucket` start full, and make `Allow` refill and then take
one token if it can.

In `NewBucket`, set `tokens` to the burst and `last` to `now()`.

In `Allow`, which already locks the mutex:

1. Refill: add `elapsed seconds × rate` tokens, but never go above `burst`.
   Set `last` to the current time.
2. If there is at least one token, take one and return `true`. Otherwise
   return `false` and take nothing.

```go
b := NewBucket(1, 2, clock.Now)
b.Allow() // => true
b.Allow() // => true
b.Allow() // => false
// clock moves forward 1s
b.Allow() // => true
```

**Think it through:**
- `burst` arrives as an `int`, but the `tokens` field is a `float64`. How
  does the struct literal already handle `burst`?
- Call `b.now()` once and use that one value for both the gap and `last`.

### 2. Reject requests with 429

**Goal:** make `Middleware(b, next)` return a handler that only calls `next`
when `b.Allow()` says yes.

Its tests use your bucket from task 1, so finish that first. For every
request:

- allowed: call `next.ServeHTTP(w, r)`;
- not allowed: don't call `next`. Set the `Retry-After` header to `"1"`,
  then use `http.Error` with status `429` and the message
  `rate limit exceeded`.

```text
POST /deploys  ->  200 deployed
POST /deploys  ->  200 deployed
POST /deploys  ->  429 rate limit exceeded      (burst 2, Retry-After: 1)
```

**Think it through:**
- What does `return` do after `http.Error`, and what happens without it?
- Which goes first: the header or the status?

### 3. Give every client its own bucket

**Goal:** make `(*PerClient).Allow(r)` use one bucket per client IP.

1. Get the IP from `r.RemoteAddr` with `net.SplitHostPort`. Add `"net"` to
   the imports.
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

The tests send requests from many goroutines at once, so `-race` catches a
map used without the lock.

**Think it through:**
- Which of the three results of `SplitHostPort` do you need?
- How does the two-value map lookup (`v, ok := m[k]`) tell you the bucket is
  missing?

## Learn more

- [Wikipedia: Token bucket](https://en.wikipedia.org/wiki/Token_bucket)
- [MDN: 429 Too Many Requests](https://developer.mozilla.org/en-US/docs/Web/HTTP/Reference/Status/429)
- [MDN: Retry-After](https://developer.mozilla.org/en-US/docs/Web/HTTP/Reference/Headers/Retry-After)
- [pkg.go.dev: net.SplitHostPort](https://pkg.go.dev/net#SplitHostPort)
