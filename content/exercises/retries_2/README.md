# Retries: retryable HTTP errors

Your deploy tool pulls container images from a registry that sometimes
answers `503` or `429`. In this exercise you fix a retry loop that tries too
often, teach it which HTTP errors are worth retrying, and write the real
sleep function it uses in production.

## Learn

### Recap

As in *Retries: backoff and jitter*, you retry only transient failures, wait
longer after each failure (`base × 2^attempt`, capped at a maximum), and stop
after `MaxAttempts` calls in total. There is no wait after the last attempt.
`Sleep` is a function in the `Policy`, so tests can pass a fake that records
the waits.

### Errors that know whether to retry

Instead of one sentinel error, any error type can carry its own answer by
having a `Retryable() bool` method. This interface describes them:

```go
type Retryable interface {
	Retryable() bool
}
```

`errors.As`, as in *Wrapping errors*, also works with an interface as the
target. It walks the chain of wrapped errors and stops at the first one that
has the interface's methods:

```go
var r Retryable
if errors.As(err, &r) {
	// r is the wrapped error, seen through the Retryable interface.
	fmt.Println(r.Retryable())
}
```

Errors from your HTTP client, DNS resolver or cloud SDK can all join in
without knowing about each other.

### Which HTTP statuses to retry

| Status | Meaning | Retry? |
|---|---|---|
| `429 Too Many Requests` | You are being rate limited. | yes, after a wait |
| `502 Bad Gateway` | The proxy couldn't reach the server. | yes |
| `503 Service Unavailable` | The server is overloaded or restarting. | yes |
| `504 Gateway Timeout` | The proxy gave up waiting for the server. | yes |
| `400`, `401`, `403`, `404`, … | Your request is wrong. | no |
| `500 Internal Server Error` | The registry team says this means a bug in their code. | no |

### Testing real waits with testing/synctest

Code that really waits, such as `time.After(3 * time.Second)`, would make
tests slow. The `testing/synctest` package runs a test inside a *bubble* with
a fake clock:

```go
synctest.Test(t, func(t *testing.T) {
	start := time.Now()
	time.Sleep(3 * time.Second) // returns instantly in real time
	fmt.Println(time.Since(start)) // 3s, exactly
})
```

Inside the bubble, when every goroutine is blocked waiting on time, the fake
clock jumps straight to the next timer. Timers fire in the right order, and
`time.Since` measures fake time exactly. Contexts with timeouts, and
`time.AfterFunc`, use the same fake clock. The tests for task 3 use this.

## Your task

### 1. Fix the retry loop

`Do` has two bugs. With `MaxAttempts: 3` and an operation that always fails,
it calls the operation 4 times and sleeps after the last one. Fix `Do` so it:

- calls `op` at most `p.MaxAttempts` times in total;
- doesn't call `Sleep` after the last attempt.

Everything else stays as it is. The tests use `Base: 1s` and `Max: 4s`:

```text
op always fails with a retryable error, MaxAttempts: 3
  calls to op:  3
  Sleep asked:  1s, 2s
  Do returns:   the error from call 3

MaxAttempts: 1
  calls to op:  1
  Sleep asked:  nothing
```

### 2. Classify retryable errors

First complete `(*StatusError).Retryable`: return `true` for codes `429`,
`502`, `503` and `504`, and `false` for every other code. A `switch` with a
multi-value `case` fits well here.

Then complete `IsRetryable(err)`. It returns `true` when `err`, or any error
it wraps, implements `Retryable` and its `Retryable()` method returns `true`.
Otherwise, including for `nil`, it returns `false`. Use `errors.As` with an
interface target, and add `"errors"` to the imports.

```go
IsRetryable(&StatusError{Code: 503})                                    // => true
IsRetryable(fmt.Errorf("pull nginx:1.27: %w", &StatusError{Code: 429})) // => true
IsRetryable(&StatusError{Code: 404})                                    // => false
IsRetryable(errors.New("disk full"))                                    // => false
```

The tests also use a DNS error type of their own with a `Retryable` method,
so check for the interface, not for `*StatusError`.

### 3. Sleep, but wake up on cancel

Complete `SleepCtx(ctx, d)`. Use a `select` that waits on two channels:

- `time.After(d)`: when it fires, return `nil`;
- `ctx.Done()`: when it closes, return `ctx.Err()`.

```go
SleepCtx(ctx, 3*time.Second) // waits 3s, returns nil

ctx, cancel := context.WithTimeout(ctx, time.Second)
SleepCtx(ctx, time.Minute) // waits 1s, returns context.DeadlineExceeded
```

In production you'd pass it as `Policy{Sleep: SleepCtx, ...}`, so that Ctrl+C
stops a retry loop straight away, instead of after a long wait.

## Learn more

- [pkg.go.dev: errors.As](https://pkg.go.dev/errors#As)
- [pkg.go.dev: testing/synctest](https://pkg.go.dev/testing/synctest)
- [Go blog: Testing concurrent code with testing/synctest](https://go.dev/blog/synctest)
- [MDN: 429 Too Many Requests](https://developer.mozilla.org/en-US/docs/Web/HTTP/Reference/Status/429)
