# Functions as values: retries and a circuit breaker

Two patterns every platform engineer meets: retrying a flaky call with
growing pauses, and a *circuit breaker* that stops calling a service that
keeps failing. Both are built from functions passed in and functions
handed back.

**New here:** a function parameter used to fake waiting; returning two closures that share state.

## Learn

### Recap

Functions are values. A parameter can be a function (`op func(attempt int) bool`),
and a function can return a function that remembers its own variables, as
in *Functions as values* and *Functions as values: state that sticks*.

### Reading a long signature

`RetryWithBackoff`'s first line is long. Split it at the commas:

| Piece | Type | Meaning |
|---|---|---|
| `attempts, baseSeconds int` | two `int`s | how many tries; the first pause |
| `op func(attempt int) bool` | a function | the call you're retrying; `true` = it worked |
| `sleep func(seconds int)` | a function with no result | "wait this many seconds" |
| `bool` (after the `)`) | result | did any attempt work? |

> **Why is it written like this?** `attempts, baseSeconds int` is Go's
> shorthand for two parameters of the same type. Go requires every part;
> keeping it on one line is the author's style, which gofmt leaves alone.

### Passing in the "sleep"

Real retry code waits between attempts. Waiting makes tests slow, so the
waiting is a parameter. A real program passes a function that really
sleeps; the tests pass one that only writes down the number it got. You just
call `sleep(...)` with the number of seconds.

### Doubling a number

To double a variable each time round a loop, change the variable itself:

```go
size := 3
size *= 2 // size is now 6
size *= 2 // size is now 12
```

`size *= 2` is short for `size = size * 2`.

### Returning two closures

A function can return two closures that share the same variables. One
changes the state, the other reads it:

```go
func NewSwitch() (func(), func() bool) {
	on := false
	flip := func() { on = !on }
	isOn := func() bool { return on }
	return flip, isOn
}

flip, isOn := NewSwitch()
flip()
isOn() // => true
```

> **Why is it written like this?** In your file, `NewCircuitBreaker`'s results
> are *named* (`record`, `isOpen`), so they already exist: assign to them
> with `=`, not `:=`. Naming them is the author's choice, to say which
> function is which.

## Your task

You're adding resilience helpers to a deploy tool.

### 1. Retry with backoff

**Goal:** make `RetryWithBackoff` call `op(1)`, `op(2)` … up to `op(attempts)`, sleeping longer before each retry, and return whether any attempt succeeded.

- Return `true` as soon as `op` returns `true`; don't call it again.
- Before each retry, call `sleep` with a wait that doubles every time:
  `baseSeconds`, then `2*baseSeconds`, then `4*baseSeconds`, and so on.
- Don't sleep before the first attempt, after a success, or after the last attempt.
- If no attempt succeeds, return `false`.

```go
// op succeeds on attempt 1
RetryWithBackoff(5, 1, op, sleep) // => true, no sleeps

// op fails three times, then succeeds on attempt 4; baseSeconds is 1
RetryWithBackoff(5, 1, op, sleep) // => true, sleeps 1, 2, 4

// op always fails; baseSeconds is 2
RetryWithBackoff(3, 2, op, sleep) // => false, sleeps 2, 4

// only one attempt, it fails
RetryWithBackoff(1, 5, op, sleep) // => false, no sleeps
```

**Think it through:**
- The wait changes over time. Where do you keep it so it survives from one loop pass to the next?
- After a failed attempt, how do you know whether another attempt is still coming?

### 2. Build a circuit breaker

**Goal:** make `NewCircuitBreaker` return `record` and `isOpen` functions that share a count of failures in a row.

- `record(ok bool)` records the result of one call: `true` is a success, `false` a failure.
- `isOpen()` returns `true` when there have been at least `threshold` failures in a row.
- A success resets the run of failures to zero, which closes the breaker.
- A new breaker is closed.

```go
record, isOpen := NewCircuitBreaker(3)
isOpen()      // => false (new breaker)
record(false)
record(false)
isOpen()      // => false (2 failures in a row)
record(false)
isOpen()      // => true (3 in a row)
record(true)
isOpen()      // => false (a success resets the run)
```

Failures with a success in between don't add up: fail, fail, success, fail,
fail leaves a threshold-3 breaker closed.

**Think it through:**
- What single number do both functions need to see? Where must it be declared so they share it?
- What should `record` do to that number for `true`, and for `false`?
- "At least `threshold`": which comparison operator is that?

## Learn more

- [Go by Example: Closures](https://gobyexample.com/closures)
- [Circuit Breaker pattern (Azure Architecture Center)](https://learn.microsoft.com/en-us/azure/architecture/patterns/circuit-breaker)
- [A Tour of Go: Function closures](https://go.dev/tour/moretypes/25)
