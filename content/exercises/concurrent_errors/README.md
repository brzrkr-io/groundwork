# Collecting errors from goroutines

A goroutine can't `return` an error to the code that started it. Yet when
you check 50 hosts at once, you need to know which ones failed, and why.
Here you collect every error, report them together, and limit how many
tasks run at once.

**New here:** one error slot per goroutine, `errors.Join` on a slice,
semaphores (a buffered channel that limits concurrency).

## Learn

### One slot per goroutine

As in *Goroutines: fix the maintenance script*, give each goroutine its own
slot in a slice and fill it by index. Errors work the same way:

```go
sizes := make([]int, len(files))
var wg sync.WaitGroup
for i, f := range files {
	wg.Go(func() {
		sizes[i] = measure(f) // each goroutine writes only its own slot
	})
}
wg.Wait() // after this, every slot is filled
```

> **Why is it written like this?** Go requires a length in `make` to assign
> `s[i]`. One slot per goroutine is convention: two goroutines writing the
> *same* variable is a data race, but separate slots are safe.

### Joining them

As in *Reporting many errors*, `errors.Join` combines errors into one. It
**skips `nil`s**, and returns `nil` if every error is `nil`. Pass a whole
slice with `...`:

```go
problems := []error{nil, errA, nil, errB}
err := errors.Join(problems...) // message: "<errA>\n<errB>"
```

The message is each error on its own line, in **slice order** (not finish
order). `errors.Is` still finds each one inside.

### Limiting concurrency with a semaphore

A *semaphore* is a counter of free slots. A buffered channel makes one:

| Action | Code | Effect |
|---|---|---|
| make | `make(chan struct{}, 3)` | 3 free slots |
| take a slot | `sem <- struct{}{}` | waits while all slots are taken |
| give it back | `<-sem` | frees a slot |

Take the slot **before** starting the goroutine, and give it back when the
goroutine ends:

```text
limit 2, 4 tasks, each takes 1s:
slot 1:  task 0 ███  task 2 ███
slot 2:  task 1 ███  task 3 ███
         0s ─────── 1s ─────── 2s
```

> **Why is it written like this?** `struct{}{}` is the empty struct value:
> it holds nothing and uses no memory. That's convention for "just a
> signal". `defer func() { <-sem }()` is required syntax to defer a
> *receive*: `defer` needs a function call, so you wrap the receive in an
> anonymous function and call it with `()`.

### Making tasks from data

A task is a `func() error`. Build one per item with a closure, as in
*Functions as values*:

```go
var jobs []func() error
for _, path := range paths {
	jobs = append(jobs, func() error {
		return os.Remove(path)
	})
}
```

Wrap an error with extra context using `fmt.Errorf` and `%w` (as in
*Errors*), so `errors.Is` still sees the original:

```go
fmt.Errorf("deleting %s: %w", path, err)
```

## Your task

You run maintenance tasks across a fleet. Every task in the tests takes 1
(fake) second, and the tests check how long your function took.

### 1. Run everything and report every error

**Goal:** `RunAll(tasks)` runs every task at the same time and returns `errors.Join` of their errors, in task order.

It returns `nil` if every task succeeded. Running them all at once means
the whole call takes 1 second, not 1 second per task. Add `"errors"` and
`"sync"` to the imports.

```go
RunAll([]func() error{ok, failsWithDiskFull, ok, failsWithTimeout})
// => an error with message "disk full\ntimeout", after 1s
```

**Think it through:**
- The starter returns on the first error. What should hold *all* the errors instead?
- What do you call on the WaitGroup before joining?

### 2. Run at most limit at once

**Goal:** `RunLimited(tasks, limit)` does what `RunAll` does, but never runs more than `limit` tasks at once.

The tests check that exactly `limit` tasks run at the same time at the
busiest moment, and the total time: 5 tasks with limit 2 take 3 seconds.

```go
RunLimited(fiveTasks, 2) // never more than 2 running; takes 3s
```

**Think it through:**
- Start from your task 1 code. How many slots does the semaphore need?
- Where in the loop do you take a slot, and where do you give it back?

### 3. Check hosts

**Goal:** `CheckHosts(hosts, limit, check)` calls `check(host)` for every host, at most `limit` at once, and returns every failure joined.

Reuse `RunLimited`; don't write the concurrency again. Each failure is
wrapped as `"<host>: <error>"`: the host name, then fixed text `": "`, then
the error, wrapped with `%w`. Healthy hosts add nothing. Add `"fmt"` to the
imports.

```go
// check fails for web-2 with "connection refused"
CheckHosts([]string{"web-1", "web-2"}, 2, check)
// => an error with message "web-2: connection refused"
```

**Think it through:**
- `RunLimited` wants `[]func() error`. What goes inside each task?
- What should a task return when `check` returns `nil`?

## Learn more

- [errors.Join](https://pkg.go.dev/errors#Join)
- [Go by Example: Errors](https://gobyexample.com/errors)
- [Effective Go: Channels as semaphores](https://go.dev/doc/effective_go#channels)
