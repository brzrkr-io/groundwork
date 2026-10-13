# Circuit breakers: protect an HTTP client

Put the breaker where it does the most good: inside an HTTP client, so every
request to a flaky backend is protected without changing the code that sends
it.

## Learn

### Recap

As in *Circuit breakers: fail fast*, a breaker is **closed** while calls
succeed, opens after `threshold` failures **in a row**, and returns `ErrOpen`
without running the call while **open**. Once the cooldown has passed it goes
**half-open** and lets one trial through: success closes it, failure opens it
again. This breaker is complete except for one bug, and it gets two new
features.

### A state-change hook

A breaker that opens silently is hard to debug at 3 a.m. You want a log line
or a metric every time it changes state. The `OnStateChange` field is a
*hook*: a function the breaker calls whenever its state changes. Callers set
it to whatever they need:

```go
b := breaker.New(5, 30*time.Second, time.Now)
b.OnStateChange = func(from, to breaker.State) {
	slog.Warn("breaker changed state", "from", from, "to", to)
}
```

Every transition already goes through one method, `setState`, so the hook only
needs calling in one place. A field holding a function is `nil` until someone
sets it, and calling a `nil` function panics, so check it first.

### http.RoundTripper

An `http.Client` doesn't send requests itself. It hands each one to its
`Transport` field, which is an `http.RoundTripper`:

```go
type RoundTripper interface {
	RoundTrip(*http.Request) (*http.Response, error)
}
```

`RoundTrip` sends one request and returns the response. `http.DefaultTransport`
is the real one that opens network connections. If you write your own type
with a `RoundTrip` method that does something and then calls a `Base`
transport, you add behaviour to **every** request the client sends. It's the
client-side version of HTTP middleware:

```go
client := &http.Client{
	Transport: &breaker.Transport{Base: http.DefaultTransport, Breaker: b},
}
resp, err := client.Get("http://inventory.internal/hosts") // goes through the breaker
```

A RoundTripper reports a failed *network* call (connection refused, timeout)
with an error. A 500 or 503 is not an error to it: the server answered, so it
returns the response and `nil`. The breaker needs to treat those 5xx answers as
failures anyway.

When `RoundTrip` returns an error, `http.Client` wraps it in a `*url.Error`
like `Get "http://...": circuit breaker is open`. `errors.Is(err, ErrOpen)`
still finds it.

The tests count requests that reach the test server with an `atomic.Int32`
from `sync/atomic`: `hits.Add(1)` and `hits.Load()` are safe to call from
several goroutines without a mutex.

## Your task

Your deploy tool talks to an inventory API that falls over during big
rollouts. Fix the breaker and wire it into the HTTP client.

### 1. Fix the failure count

The on-call team noticed the breaker opening after three timeouts spread over
a whole day, with plenty of successful calls in between. The breaker should
count *consecutive* failures. Find the bug in `record` and fix it. With a
threshold of 3:

```text
fail, ok, fail, ok, fail     => closed
fail, fail, ok, fail, fail   => closed
fail, ok, fail, fail, fail   => open
```

### 2. Call the hook on every state change

In `setState`, after updating `b.state`, call `b.OnStateChange(from, to)` when
the hook is set (not `nil`) **and** the state actually changed. A success while
already closed must not call it. The tests record each call as `"from->to"`
using the state names:

```go
// threshold 2, cooldown 30s
b.Call(fail); b.Call(fail)          // "closed->open"
// 30s later
b.Call(fail)                        // "open->half-open", "half-open->open"
// 30s later
b.Call(ok)                          // "open->half-open", "half-open->closed"
```

### 3. Send requests through the breaker

Write `Transport.RoundTrip`. Run `t.Base.RoundTrip(req)` inside
`t.Breaker.Call`, so the breaker decides whether it runs and records the
result:

- A transport error is a failure. Return it unchanged (with a `nil` response).
- A response with status code **500 or above** is a failure. Declare an
  unexported sentinel next to `Transport`,
  `var errServerError = errors.New("server error")`, and make the function you
  pass to `Call` return it. Give the caller the response and a `nil` error.
- Any other response (200, 404, …) is a success. Return it as is.
- When the breaker is open, return `nil` and the `ErrOpen` that `Call` gives
  you. The request must not reach the server.

```go
// threshold 2
client.Get(srv.URL + "/fail") // 503 response, err == nil
client.Get(srv.URL + "/fail") // 503 response, err == nil; breaker now open
client.Get(srv.URL + "/ok")   // errors.Is(err, ErrOpen); server never sees it
```

## Learn more

- [http.RoundTripper](https://pkg.go.dev/net/http#RoundTripper)
- [http.Client](https://pkg.go.dev/net/http#Client)
- [sync/atomic: Int32](https://pkg.go.dev/sync/atomic#Int32)
