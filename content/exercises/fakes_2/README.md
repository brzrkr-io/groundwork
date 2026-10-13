# Fakes: a fake clock for retries

Deploy tools retry flaky things all the time: pulling an image from a busy
registry, waiting for a load balancer to answer. A good retry waits longer
after each failure. Testing that for real would make your test sit and wait,
so this time you fake time itself.

## Learn

### Recap: inject the dependency, fake it in tests

As in *Fakes and dependency injection*, the code takes its dependency as an
interface parameter, and the test passes in a fake that records what happened.
Here the dependency is waiting:

```go
type Sleeper interface {
	Sleep(d time.Duration)
}
```

In production you pass `RealSleeper{}`, whose `Sleep` calls `time.Sleep`. In
your tests you pass a fake whose `Sleep` only appends `d` to a slice. The test
finishes instantly, and you can check every wait. Never pass `RealSleeper` in a
test: these tests would then really wait for seconds.

### Comparing durations

As in *Subtests and helpers*, compare slices with `slices.Equal`. It works on
a slice of `time.Duration` just as well as on strings:

```go
want := []time.Duration{time.Second, 2 * time.Second}
if !slices.Equal(got, want) {
	t.Errorf("slept %v, want %v", got, want) // prints [1s 2s]
}
```

To count how often `fn` is called, use a closure that changes a variable
outside it, as in *Functions as values*.

Grading works as in *Your first test*: `TestGrader` runs your two tests
against mutants of `Retry` and names any bug they missed.

## Your task

`Retry` in `exercise.go` is correct:

- it calls `fn` up to `attempts` times and returns `nil` as soon as `fn`
  returns `nil`,
- between attempts it sleeps through the `Sleeper`: `1s`, then `2s`, then
  `4s`, doubling each time,
- it never sleeps after the last attempt,
- if every attempt fails, it returns a non-nil error (fn's last error, wrapped
  as `gave up after 3 attempts: connection refused`).

```go
Retry(RealSleeper{}, 3, pullImage)
// pullImage fails, sleep 1s, pullImage fails, sleep 2s, pullImage fails
// => error "gave up after 3 attempts: ..."
```

Add the imports you need (`errors`, `slices`, `time`) to `exercise_test.go`.

### 1. Write a fake sleeper

Declare a `fakeSleeper` struct type with a `slept []time.Duration` field, and
give it a `Sleep(d time.Duration)` method with a pointer receiver that appends
`d` to `slept`. Pass it to `Retry` as `&fakeSleeper{}`.

### 2. Test a call that succeeds on the third try

In `TestRetrySucceedsOnThirdTry`, write an `fn` that returns
`errors.New("connection refused")` on its first two calls and `nil` on the
third. Call `Retry(s, 5, fn)` and check that:

- the error is `nil`,
- `fn` was called exactly 3 times (with attempts 5, a `Retry` that keeps
  going after success would call it more),
- the fake slept exactly `[1s 2s]`, compared with `slices.Equal`.

### 3. Test giving up

In `TestRetryGivesUp`, write an `fn` that always returns an error. Call
`Retry(s, 3, fn)` with a fresh fake and check that:

- the error is not `nil`,
- `fn` was called exactly 3 times,
- the fake slept exactly `[1s 2s]`, with no sleep after the last attempt.

The exercise is done when both tests pass and `TestGrader` reports that every
mutant was caught.

## Learn more

- [Package slices: Equal](https://pkg.go.dev/slices#Equal)
- [Package time: Duration](https://pkg.go.dev/time#Duration)
- [Go by Example: Closures](https://gobyexample.com/closures)
