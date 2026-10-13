# Circuit breakers: fail fast

When a dependency goes down, a *circuit breaker* stops your service from
calling it for a while. Callers get a fast error instead of a slow timeout,
and the struggling dependency gets room to recover.

**New here:** the circuit breaker states, a `String()` method for an `int`
type, an injected clock with `time.Time.Sub`.

## Learn

### Why fail fast?

Your checkout service calls a billing API with a 10-second timeout. Billing
dies. Now every checkout waits 10 seconds, requests pile up, and checkout
falls over too. Then whatever calls checkout hangs. This is a *cascading
failure*.

A breaker in front of the billing call *opens* after a few failures in a row
and rejects calls straight away, without touching the network.

### The three states

| State | What `Call` does | Leaves when |
|---|---|---|
| Closed | Runs the function. Counts failures in a row; a success resets the count to 0. | The count reaches `threshold` → Open. |
| Open | Returns `ErrOpen` at once. The function does **not** run. | `cooldown` has passed → Half-open. |
| Half-open | Lets **one** trial call run. Other calls get `ErrOpen` while it runs. | Trial succeeds → Closed. Trial fails → Open again, cooldown starts over. |

The names come from electrical breakers: "closed" means current flows,
"open" means the circuit is cut.

*In a row* matters: with a threshold of 3, fail, fail, ok, fail, fail never
opens, because the count only reaches 2.

### Naming the states

`State` is an `int` type with constants made by `iota`. A `String() string`
method lets `fmt` print a name instead of a number:

```go
type Color int

const (
	Red Color = iota
	Green
)

func (c Color) String() string {
	switch c {
	case Red:
		return "red"
	}
	return "?"
}
```

> **Why is it written like this?** The method name `String` is a Go
> convention (the `fmt.Stringer` interface); `fmt` looks for it. A fallback
> `return` after the `switch` is required: Go rejects a function that might
> end without returning.

### Failing fast with a sentinel

When open, `Call` returns the sentinel `ErrOpen`. Callers check it with
`errors.Is` and can show a friendly error or skip the feature:

```go
err := b.Call(func() error { return billing.Charge(order) })
if errors.Is(err, breaker.ErrOpen) {
	return "billing is down, try again in a minute"
}
```

### An injected clock

The breaker stores a `now func() time.Time` field instead of calling
`time.Now()` directly. Real code passes `time.Now`; tests pass a fake clock
they move by hand. `t.Sub(u)` returns how long after `u` the time `t` is, as
a `time.Duration` you can compare:

```go
age := time.Now().Sub(file.ModTime())
if age > 24*time.Hour {
	fmt.Println("stale")
}
```

### How the code is laid out

`Call` is already written. It asks `allow` whether the call may go ahead,
runs `fn`, then reports the result to `record`:

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

`allow` and `record` lock the mutex; `fn` runs **without** it, so a slow call
never blocks other callers. `State()` just returns the `state` field, so the
breaker only becomes half-open when a call arrives after the cooldown. Until
then `State()` still says `Open`.

## Your task

Your checkout service calls a flaky billing API. Finish the breaker that
protects it.

### 1. Name the states

**Goal:** make `State.String` return `"closed"`, `"open"` or `"half-open"`,
and `"unknown"` for any other value.

```go
Open.String()      // => "open"
State(7).String()  // => "unknown"
```

**Think it through:**
- Which three constants are you matching, and what is the text for each?
- What runs when none of them match?

### 2. Count failures and fail fast

**Goal:** make `record` open the breaker after `threshold` failures in a row,
and make `allow` reject calls while it is open.

In `record`, a success (`err == nil`) already resets `failures` and returns.
For a failure: add 1 to `failures`; when it reaches `threshold`, set `state`
to `Open` and `openedAt` to `b.now()`.

In `allow`, return `ErrOpen` when `state` is `Open`.

```go
b := breaker.New(3, 30*time.Second, time.Now)
b.Call(fail) // timeout error, closed (1 failure)
b.Call(fail) // timeout error, closed (2 failures)
b.Call(fail) // timeout error, now open
b.Call(fn)   // ErrOpen, fn does not run
```

**Think it through:**
- Which fields of `Breaker` change when it opens?
- The success case is already written. Your failure code goes after it.

### 3. Let a trial call through

**Goal:** make an open breaker allow one trial call after the cooldown, and
close or reopen depending on how the trial goes.

In `allow`:

- `Open` and at least `cooldown` has passed since `openedAt`: set `state` to
  `HalfOpen` and return `nil`, so the trial runs.
- `Open` and the cooldown hasn't passed: return `ErrOpen`.
- Already `HalfOpen`: a trial is running, so return `ErrOpen`.

In `record`:

- A success sets `state` to `Closed` (as well as resetting `failures`).
- A failure while `HalfOpen` sets `state` to `Open` and `openedAt` to
  `b.now()`, so the cooldown starts over.

```go
// opened at 09:00:00 with a 30s cooldown
// 09:00:29  b.Call(fn) => ErrOpen
// 09:00:30  b.Call(fn) => fn runs; State() is HalfOpen while it runs
//           fn succeeds => State() is Closed
```

**Think it through:**
- Exactly 30s counts as "passed". Which comparison says "not passed yet"?
- A failed trial must reopen even if `failures` is below `threshold`. Which
  condition in `record` needs a second part?

## Learn more

- [Martin Fowler: CircuitBreaker](https://martinfowler.com/bliki/CircuitBreaker.html)
- [Microsoft: Circuit Breaker pattern](https://learn.microsoft.com/en-us/azure/architecture/patterns/circuit-breaker)
- [pkg.go.dev: time.Time.Sub](https://pkg.go.dev/time#Time.Sub)
- [pkg.go.dev: fmt.Stringer](https://pkg.go.dev/fmt#Stringer)
