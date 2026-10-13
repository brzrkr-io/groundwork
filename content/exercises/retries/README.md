# Retries: backoff and jitter

Networks drop packets, load balancers restart and databases fail over. A
service that tries again, politely, rides out these blips instead of paging
you at 3am. In this exercise you build a small retry helper of the kind every
deploy tool and API client carries.

## Learn

### Transient or permanent?

Only retry failures that might go away on their own. A *transient* failure
is a blip. A *permanent* failure gives the same answer no matter how often
you ask.

| Transient: retry | Permanent: give up straight away |
|---|---|
| Request timed out | `400 Bad Request` (your request is wrong) |
| `503 Service Unavailable` | `404 Not Found` (it isn't there) |
| Connection reset by peer | Bad credentials (`401`/`403`) |

Retrying a permanent failure wastes time and hides the real problem.

In this exercise, an operation marks a transient failure by wrapping the
sentinel error `ErrTemporary` with `%w`, as in *Wrapping errors*:

```go
return fmt.Errorf("GET /healthz: 503: %w", ErrTemporary)
```

so `errors.Is(err, ErrTemporary)` tells you whether to try again. Any other
error is permanent.

### Exponential backoff

If a service is struggling, hammering it makes things worse. So you wait
between attempts, and you wait longer each time. *Exponential backoff* doubles
the wait after every failure: `base × 2^attempt`.

| attempt | wait with base = 100ms |
|---|---|
| 0 | 100ms |
| 1 | 200ms |
| 2 | 400ms |
| 3 | 800ms |

Doubling grows fast, so you also set a *cap*: the wait never goes above some
maximum, such as 2 seconds.

`time.Duration` is a whole number of nanoseconds, so you can multiply it:

```go
d := 100 * time.Millisecond
d *= 2 // 200ms
```

### Jitter and the thundering herd

Picture 500 pods that all lose the database at the same moment. With plain
backoff they all retry at exactly 100ms, then all at 200ms, and so on. Each
wave hits the database at once. This is the *thundering herd*. *Jitter* adds
randomness so the retries spread out.

You will use *equal jitter*: keep half of the wait, and randomise the other
half.

```text
jittered = d/2 + rnd × d/2        where 0 ≤ rnd < 1
```

The random number comes from a function `rnd func() float64`, such as
`rand.Float64` from `math/rand/v2`. Tests pass a function that always returns
the same number, so the result is predictable. Multiplying a `Duration` by a
fraction means converting to `float64` and back:

```go
half := d / 2
extra := time.Duration(0.5 * float64(half)) // a quarter of d
```

### Stopping

A retry loop needs limits:

- **Max attempts.** `MaxAttempts: 3` means at most 3 calls in total: the
  first try plus 2 retries. After the last attempt there is nothing left to
  wait for, so don't sleep.
- **Cancellation.** If the caller's context is cancelled (the user pressed
  Ctrl+C, or a deadline passed), stop waiting and return.

### Why sleeping is injected

Real code waits with something like `time.After` inside a `select` on
`ctx.Done()`. Tests can't wait seconds for every case. So `Policy` takes a
`Sleep` function. Production passes a real one; the tests pass a fake that
just writes down each requested duration and returns straight away. That
makes the tests instant and exact.

## Your task

Your deploy tool calls the cluster API, which sometimes answers `503` while it
restarts. Build the retry helper in `exercise.go`.

### 1. Compute the backoff

Complete `Backoff(attempt, base, limit)`. Return `base` doubled `attempt`
times, but never more than `limit`. Attempt `0` returns `base`.

```go
Backoff(0, 100*time.Millisecond, 2*time.Second) // => 100ms
Backoff(3, 100*time.Millisecond, 2*time.Second) // => 800ms
Backoff(5, 100*time.Millisecond, 2*time.Second) // => 2s (3.2s is over the limit)
```

### 2. Add jitter

Complete `Jitter(d, rnd)`. Call `rnd()` once and return
`d/2 + rnd() × d/2`.

```go
Jitter(2*time.Second, func() float64 { return 0 })          // => 1s
Jitter(2*time.Second, func() float64 { return 0.5 })        // => 1.5s
Jitter(400*time.Millisecond, func() float64 { return 0.25 }) // => 250ms
```

### 3. Retry the operation

Complete `Do(ctx, p, op)`. The loop that calls `op` up to `p.MaxAttempts`
times is there. Change it to `for attempt := range p.MaxAttempts` so you have
the attempt number (`0` for the first call). Then, after each call:

1. If `op` returned `nil`, return `nil`.
2. If the error is permanent (`errors.Is(err, ErrTemporary)` is false), return
   that error straight away.
3. If this was the last attempt, stop and return the error from `op`. Don't
   sleep.
4. Otherwise call `p.Sleep(ctx, wait)` with
   `wait = Jitter(Backoff(attempt, p.Base, p.Max), p.Rand)`. If `Sleep` returns
   an error, return that error.

The tests use `Base: 100ms`, `Max: 800ms` and a `Rand` that always returns
`0.5`, so the waits are `75ms`, `150ms`, `300ms`, `600ms`, `600ms`, …

```text
op fails with a temporary error twice, then succeeds (MaxAttempts: 5)
  calls to op:  3
  Sleep asked:  75ms, 150ms
  Do returns:   nil

op always fails with "call N: temporary failure" (MaxAttempts: 3)
  calls to op:  3
  Sleep asked:  75ms, 150ms   (no sleep after the last call)
  Do returns:   the error from call 3
```

## Learn more

- [AWS Architecture Blog: Exponential backoff and jitter](https://aws.amazon.com/blogs/architecture/exponential-backoff-and-jitter/)
- [pkg.go.dev: errors.Is](https://pkg.go.dev/errors#Is)
- [pkg.go.dev: time.Duration](https://pkg.go.dev/time#Duration)
- [pkg.go.dev: math/rand/v2 Float64](https://pkg.go.dev/math/rand/v2#Float64)
