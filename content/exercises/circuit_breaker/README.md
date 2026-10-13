# Circuit breakers: fail fast

When a dependency goes down, a *circuit breaker* stops your service from
calling it for a while. Callers get a fast error instead of a slow timeout,
and the struggling dependency gets room to recover.

## Learn

### Why fail fast?

Say your checkout service calls a billing API with a 10-second timeout. The
billing API dies. Now every checkout request waits 10 seconds before it fails.
Requests pile up, goroutines and connections run out, and checkout falls over
too. Anything that calls checkout starts hanging next. One dead service has
taken down three. This is a *cascading failure*.

A circuit breaker sits in front of the billing call. After a few failures in a
row it *opens* and rejects calls straight away, without touching the network.
Failing in a microsecond is much better than failing in 10 seconds.

### The three states

```text
              threshold failures in a row
   ┌────────┐ ───────────────────────────▶ ┌──────┐
   │ closed │                              │ open │ ◀─┐
   └────────┘ ◀──┐                         └──────┘   │
                 │ trial succeeds            │        │ trial fails
                 │                  cooldown │        │
                 │                    passed ▼        │
                 │                     ┌───────────┐  │
                 └──────────────────── │ half-open │ ─┘
                                       └───────────┘
```

| State | What `Call` does | Leaves when |
|---|---|---|
| Closed | Runs the function. Counts failures in a row. A success resets the count to 0. | The count reaches the threshold → Open. |
| Open | Returns `ErrOpen` straight away. The function does **not** run. | The cooldown has passed → Half-open. |
| Half-open | Lets **one** trial call run. Other calls get `ErrOpen` while it runs. | Trial succeeds → Closed (count reset). Trial fails → Open again, and the cooldown starts over. |

The name comes from electrical circuit breakers: "closed" means current flows
(calls go through), "open" means the circuit is cut.

### Counting *consecutive* failures

The breaker counts failures **in a row**. One success sets the count back to
0. With a threshold of 3, the sequence fail, fail, ok, fail, fail never opens
the breaker, because the count only ever reaches 2.

### Failing fast with a sentinel

When the breaker is open, `Call` returns the sentinel error `ErrOpen`. Callers
check for it with `errors.Is` and can show a cached page, a friendly error, or
skip the feature:

```go
err := b.Call(func() error { return billing.Charge(order) })
if errors.Is(err, breaker.ErrOpen) {
	return "billing is down, try again in a minute"
}
```

### An injected clock

The breaker needs the time to know when the cooldown has passed. Instead of
calling `time.Now()` directly, it stores a `now func() time.Time` field. Real
code passes `time.Now`, and the tests pass a fake clock that they move forward
by hand. `t.Sub(u)` returns how long after `u` the time `t` is:

```go
elapsed := b.now().Sub(b.openedAt) // a time.Duration
if elapsed < b.cooldown {
	// still cooling down
}
```

### How the code is laid out

`Call` is already written. It asks `allow` whether the call may go ahead, runs
`fn`, and then reports the result to `record`:

```go
func (b *Breaker) Call(fn func() error) error {
	if err := b.allow(); err != nil {
		return err
	}
	err := fn()
	b.record(err)
	return err
}
```

`allow` and `record` lock the mutex, so many goroutines can share one breaker
safely. `fn` runs **without** the lock held, so a slow call never blocks other
callers. The breaker stores its state in the `state` field, and `State()` just
returns it. That means the breaker only becomes half-open when a call arrives
after the cooldown; until then `State()` still reports `Open`.

## Your task

Your checkout service calls a flaky billing API. Finish the breaker that
protects it.

### 1. Name the states

Make `State.String` return `"closed"`, `"open"` and `"half-open"` for the three
states, and `"unknown"` for any other value:

```go
Open.String()      // => "open"
State(7).String()  // => "unknown"
```

### 2. Count failures and fail fast

In `record`, a success (`err == nil`) already sets `failures` back to 0 and
returns. Handle a failure: add 1 to `failures`, and when it reaches
`threshold`, set `state` to `Open` and set `openedAt` to `b.now()`.

In `allow`, return `ErrOpen` when `state` is `Open`.

```go
b := breaker.New(3, 30*time.Second, time.Now)
b.Call(fail) // timeout error, closed (1 failure)
b.Call(fail) // timeout error, closed (2 failures)
b.Call(fail) // timeout error, now open
b.Call(fn)   // ErrOpen, fn does not run
```

### 3. Let a trial call through

In `allow`, when `state` is `Open` and at least `cooldown` has passed since
`openedAt`, set `state` to `HalfOpen` and return `nil` so the trial runs. When
`state` is already `HalfOpen`, a trial is running, so return `ErrOpen`.

In `record`, a success sets `state` to `Closed` (as well as resetting
`failures`). A failure while `HalfOpen` sets `state` to `Open` and `openedAt`
to `b.now()`, so the cooldown starts over.

```go
// opened at 09:00:00 with a 30s cooldown
// 09:00:29  b.Call(fn) => ErrOpen
// 09:00:30  b.Call(fn) => fn runs; State() is HalfOpen while it runs
//           fn succeeds => State() is Closed
```

## Learn more

- [Martin Fowler: CircuitBreaker](https://martinfowler.com/bliki/CircuitBreaker.html)
- [Microsoft: Circuit Breaker pattern](https://learn.microsoft.com/en-us/azure/architecture/patterns/circuit-breaker)
- [time.Time.Sub](https://pkg.go.dev/time#Time.Sub)
- [sync.Mutex](https://pkg.go.dev/sync#Mutex)
