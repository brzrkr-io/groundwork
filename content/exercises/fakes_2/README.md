# Fakes: a fake clock for retries

Deploy tools retry flaky things all the time: pulling an image from a busy
registry, waiting for a load balancer to answer. A good retry waits longer
after each failure. Testing that for real would make your test sit and wait,
so this time you fake time itself.

**New here:** faking time, counting calls with a closure, comparing `time.Duration` slices.

## Learn

### Recap: inject the dependency, fake it in tests

As in *Fakes and dependency injection*, the code takes its dependency as an
interface parameter, and the test passes in a fake that records what
happened. Here the dependency is *waiting*:

```go
type Sleeper interface {
	Sleep(d time.Duration)
}
```

In production you pass `RealSleeper{}`, whose `Sleep` calls `time.Sleep`.
In your tests you pass a fake that only records `d`. The test finishes
instantly, and you can check every wait.

Never pass `RealSleeper` in a test: it would really wait for seconds.

### Counting calls with a closure

A *closure* is a function literal that uses a variable from outside it, as in
*Functions as values*. It can change that variable, and the change sticks:

```go
pings := 0
ping := func() bool {
	pings++
	return pings > 1 // false the first time, true after that
}
ping()
ping()
// pings is now 2
```

> **Why is it written like this?** `func() bool { ... }` is a function
> literal; Go requires the `func` keyword, the parameter list (empty here)
> and the result type, the same parts as a named function minus the name.
> The counter must be declared *before* the closure so the closure can see
> it. The names `pings` and `ping` are your choice.

### Comparing durations

As in *Subtests and helpers*, compare slices with `slices.Equal`. It works on
`[]time.Duration` just like on strings:

```go
want := []time.Duration{500 * time.Millisecond, time.Minute}
if !slices.Equal(got, want) {
	t.Errorf("waited %v, want %v", got, want) // prints [500ms 1m0s]
}
```

`time.Second` is one second; `3 * time.Second` is three.

Grading works as in *Your first test*: `TestGrader` runs your two tests
against *mutants* of `Retry` and names any bug they missed.

## Your task

`Retry(s, attempts, fn)` in `exercise.go` is correct:

- it calls `fn` up to `attempts` times and returns `nil` as soon as `fn`
  returns `nil`,
- between attempts it sleeps through the `Sleeper` `s`: `1s`, then `2s`, then
  `4s`, doubling each time,
- it never sleeps after the last attempt,
- if every attempt fails, it returns a non-nil error (fn's last error,
  wrapped as `gave up after 3 attempts: connection refused`).

```go
Retry(RealSleeper{}, 3, pullImage)
// pullImage fails, sleep 1s, pullImage fails, sleep 2s, pullImage fails
// => error "gave up after 3 attempts: ..."
```

Add the imports you need (`errors`, `slices`, `time`) to `exercise_test.go`.

### 1. Write a fake sleeper

**Goal:** declare a `fakeSleeper` struct with a field `slept
[]time.Duration`, and a `Sleep` method that appends each duration to it
instead of waiting.

Use a pointer receiver, and pass the fake to `Retry` with `&` in front.

**Think it through:**
- What exact method does `Sleeper` require? Match its name and parameter.
- Which shape from *Fakes and dependency injection* can you reuse?

### 2. Test a call that succeeds on the third try

**Goal:** in `TestRetrySucceedsOnThirdTry`, check that `Retry(s, 5, fn)`
stops as soon as `fn` succeeds.

Write an `fn` that returns `errors.New("connection refused")` on its first
two calls and `nil` on the third. Then check:

- the error is `nil`,
- `fn` was called exactly 3 times (a `Retry` that keeps going after success
  would call it more),
- the fake slept exactly `[1s 2s]`, compared with `slices.Equal`.

**Think it through:**
- Where does the call counter live so that both `fn` and your checks can see
  it?
- What does the counter's value tell `fn` about whether to fail or succeed?

### 3. Test giving up

**Goal:** in `TestRetryGivesUp`, check that `Retry(s, 3, fn)` with an `fn`
that always fails gives up properly.

Use a fresh fake, then check:

- the error is **not** `nil`,
- `fn` was called exactly 3 times,
- the fake slept exactly `[1s 2s]`, with no sleep after the last attempt.

**Think it through:**
- What changes from task 2: the `fn`, the attempts, and which error result
  counts as correct?

The exercise is done when both tests pass and `TestGrader` reports that every
mutant was caught.

## Learn more

- [Package slices: Equal](https://pkg.go.dev/slices#Equal)
- [Package time: Duration](https://pkg.go.dev/time#Duration)
- [Go by Example: Closures](https://gobyexample.com/closures)
