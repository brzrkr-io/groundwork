# Functions as values: retries and a circuit breaker

Two patterns every platform engineer meets: retrying a flaky call with
growing pauses, and a *circuit breaker* that stops calling a service that
keeps failing. Both are built from functions passed in and functions
handed back.

## Learn

### Recap

Functions are values: a parameter can be a function (`op func(attempt int) bool`),
and a function can return a function that remembers its own variables, as
in *Functions as values* and *Functions as values: state that sticks*.

### Passing in the "sleep"

Real retry code waits between attempts. Waiting makes tests slow, so
`RetryWithBackoff` takes the waiting as a parameter: `sleep func(seconds int)`.
A real program would pass a function that really sleeps. The tests pass one
that only writes down the number it got.

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

## Your task

### 1. Retry with backoff

Write `RetryWithBackoff`. Call `op(1)`, `op(2)`, ... up to `op(attempts)`,
and return `true` as soon as one returns `true`. Before each retry, call
`sleep` with a wait that doubles every time: `baseSeconds`, then
`2*baseSeconds`, then `4*baseSeconds`, and so on. Don't sleep before the
first attempt, after a success, or after the last attempt. If no attempt
succeeds, return `false`.

```go
// op fails three times, then succeeds on attempt 4; baseSeconds is 1
RetryWithBackoff(5, 1, op, sleep) // => true, sleeps 1, 2, 4

// op always fails; baseSeconds is 2
RetryWithBackoff(3, 2, op, sleep) // => false, sleeps 2, 4
```

### 2. Build a circuit breaker

Write `NewCircuitBreaker`. It returns two functions that share state:

- `record(ok bool)` records the result of one call to a service.
- `isOpen()` returns `true` when the last `threshold` recorded results
  were all failures (`false`), in other words, when there have been at
  least `threshold` failures in a row.

A success resets the run of failures, so it closes the breaker again. A new
breaker is closed.

```go
record, isOpen := NewCircuitBreaker(3)
record(false)
record(false)
isOpen() // => false (2 failures in a row)
record(false)
isOpen() // => true (3 in a row)
record(true)
isOpen() // => false (a success resets the run)
```

## Learn more

- [Go by Example: Closures](https://gobyexample.com/closures)
- [Circuit Breaker pattern (Azure Architecture Center)](https://learn.microsoft.com/en-us/azure/architecture/patterns/circuit-breaker)
- [A Tour of Go: Function closures](https://go.dev/tour/moretypes/25)
