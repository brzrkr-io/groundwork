# Circuit breakers: protect an HTTP client

Put the breaker where it does the most good: inside an HTTP client, so every
request to a flaky backend is protected without changing the code that
sends it.

**New here:** a hook field (a function stored in a struct), `http.RoundTripper`,
capturing a result from a closure.

## Learn

### Recap

As in *Circuit breakers: fail fast*, a breaker is **closed** while calls
succeed, opens after `threshold` failures **in a row**, and returns
`ErrOpen` without running the call while **open**. After the cooldown it
goes **half-open** and lets one trial through: success closes it, failure
opens it again. This breaker is complete except for one bug, and it gets two
new features.

### A state-change hook

A breaker that opens silently is hard to debug at 3 a.m. The `OnStateChange`
field is a *hook*: a function the breaker calls whenever its state changes.
Callers set it to whatever they need:

```go
b := breaker.New(5, 30*time.Second, time.Now)
b.OnStateChange = func(from, to breaker.State) {
	slog.Warn("breaker changed state", "from", from, "to", to)
}
```

Every transition goes through one method, `setState`, so the hook is called
in one place.

> **Why is it written like this?** Go requires care here: a function field
> is `nil` until someone sets it, and calling a `nil` function panics. So
> check `!= nil` first. Calling the field looks like calling a method:
> `b.OnStateChange(x, y)`.

### http.RoundTripper

An `http.Client` doesn't send requests itself. It hands each one to its
`Transport` field, an `http.RoundTripper`:

```go
type RoundTripper interface {
	RoundTrip(*http.Request) (*http.Response, error)
}
```

`http.DefaultTransport` is the real one that opens connections. Your own
type with a `RoundTrip` method can do something, then call a `Base`
transport. That adds behaviour to **every** request the client sends, like
middleware on the client side:

```go
client := &http.Client{
	Transport: &breaker.Transport{Base: http.DefaultTransport, Breaker: b},
}
resp, err := client.Get("http://inventory.internal/hosts") // goes through the breaker
```

A RoundTripper reports a failed *network* call (connection refused, timeout)
with an error. A 500 or 503 is **not** an error to it: the server answered.
The breaker must treat those 5xx answers as failures anyway.

When `RoundTrip` returns an error, `http.Client` wraps it in a `*url.Error`
like `Get "http://...": circuit breaker is open`. `errors.Is(err, ErrOpen)`
still finds it.

### Getting a value out of a closure

`Call` takes a `func() error`, so the function you pass can only return an
error. To keep anything else, declare a variable *outside* and assign it
*inside*. The closure can see and change it:

```go
var body string
err := retry(func() error {
	var err error
	body, err = fetch()
	return err
})
```

> **Why is it written like this?** `=` (not `:=`) inside the closure is what
> Go needs to assign the outer `body`; `:=` would create a new local one.

The tests count requests that reach the test server with an `atomic.Int32`
(`hits.Add(1)`, `hits.Load()`), safe across goroutines without a mutex.

## Your task

Your deploy tool talks to an inventory API that falls over during big
rollouts. Fix the breaker and wire it into the HTTP client.

### 1. Fix the failure count

**Goal:** make `record` count *consecutive* failures only.

The on-call team saw the breaker open after three timeouts spread over a
whole day, with plenty of successes in between. Find the bug in `record` and
fix it. With a threshold of 3:

```text
fail, ok, fail, ok, fail     => closed
fail, fail, ok, fail, fail   => closed
fail, ok, fail, fail, fail   => open
```

**Think it through:**
- Look at the success branch. What should a success do to `failures`?

### 2. Call the hook on every state change

**Goal:** make `setState` call `b.OnStateChange(from, to)` whenever the state
really changes.

Call it after updating `b.state`, only when the hook is set (not `nil`)
**and** the old state differs from `to`. A success while already closed must
not call it. The tests record each call as `"from->to"` using the state
names:

```go
// threshold 2, cooldown 30s
b.Call(fail); b.Call(fail)          // "closed->open"
// 30s later
b.Call(fail)                        // "open->half-open", "half-open->open"
// 30s later
b.Call(ok)                          // "open->half-open", "half-open->closed"
```

**Think it through:**
- `setState` only receives `to`. Where is the old state, and when must you
  save it?
- Two conditions must both be true. Which operator joins them?

### 3. Send requests through the breaker

**Goal:** make `Transport.RoundTrip` send each request inside
`t.Breaker.Call`, counting transport errors and 5xx responses as failures.

Run `t.Base.RoundTrip(req)` inside the function you pass to `t.Breaker.Call`:

- A transport error is a failure. Return it unchanged (with a `nil`
  response).
- A status code of **500 or above** is a failure. Declare an unexported
  sentinel next to `Transport` named `errServerError` with the message
  `server error`, and return it from the function you pass to `Call`. The
  caller still gets the response and a `nil` error.
- Any other response (200, 404, …) is a success. Return it as is.
- When the breaker is open, return a `nil` response and the `ErrOpen` that
  `Call` gives you. The request must not reach the server.

```go
// threshold 2
client.Get(srv.URL + "/fail") // 503 response, err == nil
client.Get(srv.URL + "/fail") // 503 response, err == nil; breaker now open
client.Get(srv.URL + "/ok")   // errors.Is(err, ErrOpen); server never sees it
```

**Think it through:**
- The response is made inside the closure but returned outside it. Where
  must its variable be declared?
- After `Call` returns, which error means "5xx, hand back the response", and
  how do you check for it?

## Learn more

- [pkg.go.dev: http.RoundTripper](https://pkg.go.dev/net/http#RoundTripper)
- [pkg.go.dev: http.Client](https://pkg.go.dev/net/http#Client)
- [pkg.go.dev: sync/atomic Int32](https://pkg.go.dev/sync/atomic#Int32)
