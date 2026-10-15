# Retries: backoff and jitter

Networks drop packets, load balancers restart and databases fail over. A
service that tries again, politely, rides out these blips instead of paging
you at 3am. Here you build the retry helper every deploy tool carries.

**New here:** exponential backoff, jitter, multiplying a `time.Duration`.

## Learn

### Transient or permanent?

Only retry failures that might go away on their own.

| Transient: retry | Permanent: give up straight away |
|---|---|
| Request timed out | `400 Bad Request` (your request is wrong) |
| `503 Service Unavailable` | `404 Not Found` (it isn't there) |
| Connection reset by peer | Bad credentials (`401`/`403`) |

In this exercise an operation marks a transient failure by wrapping the
sentinel error `ErrTemporary` with `%w`, as in *Wrapping errors*. So
`errors.Is(err, ErrTemporary)` tells you whether to try again. Any other error
is permanent.

### Exponential backoff

Hammering a struggling service makes it worse, so you wait longer after each
failure. *Exponential backoff* doubles the wait every time:

| attempt | wait with base = 100ms |
|---|---|
| 0 | 100ms |
| 1 | 200ms |
| 2 | 400ms |
| 3 | 800ms |

Doubling grows fast, so a *cap* (limit) keeps the wait from going above a
maximum, such as 2 seconds.

`time.Duration` is a whole number of nanoseconds, so you can multiply it:

```go
timeout := 3 * time.Second
timeout *= 10 // 30s
```

> **Why is it written like this?**
> - **Go requires** both sides of `*` to have the same type. `3` is an untyped
>   constant, so it becomes a `Duration`; an `int` *variable* would need
>   `time.Duration(n)`.
> - **Convention:** `x *= 2` is short for `x = x * 2`.

### Jitter and the thundering herd

If 500 pods lose the database at once, plain backoff makes them all retry at
exactly 100ms, then 200ms… Each wave hits at once: the *thundering herd*.
*Jitter* adds randomness to spread them out.

You'll use *equal jitter*: keep half the wait, randomise the other half.

```text
jittered = d/2 + rnd × d/2        where 0 ≤ rnd < 1
```

The random number comes from a function `rnd func() float64`. Tests pass one
that always returns the same number. To multiply a `Duration` by a fraction,
convert to `float64` and back:

```go
third := time.Duration(0.33 * float64(time.Minute)) // about 20s
```

> **Why is it written like this?**
> - **Go requires** the conversions: Go never mixes `float64` and `Duration`
>   by itself.
> - **Author's choice:** the parameter name `rnd`.

### Stopping, and why sleep is injected

- **Max attempts.** `MaxAttempts: 3` means at most 3 calls: the first try
  plus 2 retries. After the last attempt there's nothing to wait for, so
  don't sleep.
- **Cancellation.** `p.Sleep` returns an error if the context is cancelled
  while waiting. Stop and return it.

`Policy` takes a `Sleep` function so tests don't wait for real. The fake just
records each duration and returns at once.

## Your task

Your deploy tool calls the cluster API, which sometimes answers `503` while it
restarts. Build the retry helper in `exercise.go`.

### 1. Compute the backoff

**Goal:** make `Backoff(attempt, base, limit)` return `base` doubled once per
attempt, but never more than `limit`.

Attempt `0` returns `base`.

```go
Backoff(0, 100*time.Millisecond, 2*time.Second) // => 100ms
Backoff(3, 100*time.Millisecond, 2*time.Second) // => 800ms
Backoff(5, 100*time.Millisecond, 2*time.Second) // => 2s (3.2s is over the limit)
```

**Think it through:**
- Which loop from *Looping with range* runs exactly `attempt` times?
- When should you return `limit` instead? Checking inside the loop also stops
  huge attempt numbers from overflowing.

### 2. Add jitter

**Goal:** make `Jitter(d, rnd)` return `d/2 + rnd() × d/2`.

Call `rnd()` once.

```go
Jitter(2*time.Second, func() float64 { return 0 })           // => 1s
Jitter(2*time.Second, func() float64 { return 0.5 })         // => 1.5s
Jitter(400*time.Millisecond, func() float64 { return 0.25 }) // => 250ms
```

**Think it through:**
- `rnd` is a function: you need its *result*, `rnd()`, not its name.
- Which part needs converting to `float64` and back to `time.Duration`?

### 3. Retry the operation

**Goal:** make `Do(ctx, p, op)` retry temporary failures with jittered
backoff, and stop on success, a permanent error, the last attempt, or a
`Sleep` error.

Change the loop to `for attempt := range p.MaxAttempts` so you have the attempt
number (`0` for the first call). After each call to `op`:

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

**Think it through:**
- `attempt` counts from 0. What is its value on the last attempt?
- Keep `Sleep`'s error in its own variable so it doesn't overwrite `err`.

## Learn more

- [AWS Architecture Blog: Exponential backoff and jitter](https://aws.amazon.com/blogs/architecture/exponential-backoff-and-jitter/)
- [pkg.go.dev: errors.Is](https://pkg.go.dev/errors#Is)
- [pkg.go.dev: time.Duration](https://pkg.go.dev/time#Duration)
- [pkg.go.dev: math/rand/v2 Float64](https://pkg.go.dev/math/rand/v2#Float64)
