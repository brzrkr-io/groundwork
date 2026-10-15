# Retries: retryable HTTP errors

Your deploy tool pulls container images from a registry that sometimes
answers `503` or `429`. You fix a retry loop that tries too often, teach it
which HTTP errors are worth retrying, and write the real sleep it uses.

**New here:** `errors.As` with an interface target, `select` on a timer and a
context, `testing/synctest`.

## Learn

### Recap

As in *Retries: backoff and jitter*, you retry only transient failures and
wait longer after each one (`base × 2^attempt`, capped at `Max`). You stop
after `MaxAttempts` calls **in total**, and there is no wait after the last
call.

`Sleep` is a field of `Policy` that holds a function. Tests put a fake there
that records the waits instead of really waiting.

### Errors that know whether to retry

Any error type can carry its own answer by having a `Retryable() bool`
method. This interface describes all such errors:

```go
type Retryable interface {
	Retryable() bool
}
```

`errors.As`, as in *Wrapping errors*, also accepts a pointer to an interface
variable. It walks the chain of wrapped errors and stops at the first one
that has the interface's methods:

```go
var t interface{ Timeout() bool }
if errors.As(err, &t) {
	fmt.Println("timeout?", t.Timeout())
}
```

`errors.As` returns `true` only if it *found* one. It doesn't call the
method for you; you still have to ask it.

> **Why is it written like this?**
> - **Go requires** the `&`: `errors.As` needs a pointer so it can fill the
>   variable in.
> - **Convention:** a one-method interface is named after what it does
>   (`Retryable`, `Timeout`).
> - **Author's choice:** checking for the interface, not one concrete type,
>   so errors from your HTTP client, DNS resolver or cloud SDK can all join
>   in.

### Which HTTP statuses to retry

| Status | Meaning | Retry? |
|---|---|---|
| `429 Too Many Requests` | You are being rate limited. | yes |
| `502 Bad Gateway` | The proxy couldn't reach the server. | yes |
| `503 Service Unavailable` | The server is overloaded or restarting. | yes |
| `504 Gateway Timeout` | The proxy gave up waiting for the server. | yes |
| `400`, `401`, `403`, `404`, … | Your request is wrong. | no |
| `500 Internal Server Error` | The registry team says this means a bug in their code. | no |

A `case` in a `switch` can list several values, separated by commas:

```go
switch day {
case "Sat", "Sun":
	return "weekend"
}
```

### Waiting for whichever happens first

A `select` waits on several channels and runs the `case` of the first one
that is ready:

```go
select {
case msg := <-inbox:
	fmt.Println(msg)
case <-quit:
	fmt.Println("bye")
}
```

`time.After(d)` returns a channel that receives once `d` has passed.
`ctx.Done()` returns a channel that closes when the context is cancelled or
times out, and then `ctx.Err()` says why.

> **Why is it written like this?**
> - **Go requires** `<-ch` in a `case` to receive from a channel.
> - **Author's choice:** the order of the cases. It doesn't change which one
>   wins.

### Testing real waits with testing/synctest

The tests for task 3 run inside a `synctest` *bubble* with a fake clock.
When every goroutine is waiting on time, the clock jumps straight to the next
timer. A 3-second wait finishes instantly, and `time.Since` still measures
exactly 3s. Timers and context timeouts both use this clock, so you write
normal code.

## Your task

### 1. Fix the retry loop

**Goal:** make `Do` call `op` at most `p.MaxAttempts` times and never sleep
after the last attempt.

`Do` has two bugs. With `MaxAttempts: 3` and an operation that always fails,
it calls `op` 4 times and sleeps after the last one. Everything else stays
as it is. The tests use `Base: 1s` and `Max: 4s`:

```text
op always fails with a retryable error, MaxAttempts: 3
  calls to op:  3
  Sleep asked:  1s, 2s
  Do returns:   the error from call 3

MaxAttempts: 1
  calls to op:  1
  Sleep asked:  nothing
```

**Think it through:**
- `attempt` starts at 0. Which comparison in the `for` line lets it run one
  time too many?
- After `op` fails, how can you tell this was the last attempt, before
  calling `Sleep`?

### 2. Classify retryable errors

**Goal:** make `(*StatusError).Retryable` and `IsRetryable` say whether an
error is worth retrying.

`(*StatusError).Retryable` returns `true` for codes `429`, `502`, `503` and
`504`, and `false` for every other code.

`IsRetryable(err)` returns `true` when `err`, or any error it wraps,
implements `Retryable` **and** its `Retryable()` method returns `true`.
Otherwise, including for `nil`, it returns `false`. Add `"errors"` to the
imports.

```go
IsRetryable(&StatusError{Code: 503})                                    // => true
IsRetryable(fmt.Errorf("pull nginx:1.27: %w", &StatusError{Code: 429})) // => true
IsRetryable(&StatusError{Code: 404})                                    // => false
IsRetryable(errors.New("disk full"))                                    // => false
```

The tests also use a DNS error type of their own with a `Retryable` method.

**Think it through:**
- Which field of `StatusError` holds the code, and which four values say yes?
- What type should the variable you pass to `errors.As` have, so that the
  DNS error matches too?
- Finding a `Retryable` isn't enough. What else must be true?

### 3. Sleep, but wake up on cancel

**Goal:** make `SleepCtx(ctx, d)` wait for `d`, but return early when `ctx`
is done.

- When `time.After(d)` fires first, return `nil`.
- When `ctx.Done()` closes first, return `ctx.Err()`.

```go
SleepCtx(ctx, 3*time.Second) // waits 3s, returns nil

ctx, cancel := context.WithTimeout(ctx, time.Second)
SleepCtx(ctx, time.Minute) // waits 1s, returns context.DeadlineExceeded
```

In production you'd pass it as `Policy{Sleep: SleepCtx, ...}`, so Ctrl+C
stops a retry loop straight away instead of after a long wait.

**Think it through:**
- Which two channels are you waiting on, and what does each case return?

## Learn more

- [pkg.go.dev: errors.As](https://pkg.go.dev/errors#As)
- [pkg.go.dev: testing/synctest](https://pkg.go.dev/testing/synctest)
- [Go blog: Testing concurrent code with testing/synctest](https://go.dev/blog/synctest)
- [MDN: 429 Too Many Requests](https://developer.mozilla.org/en-US/docs/Web/HTTP/Reference/Status/429)
