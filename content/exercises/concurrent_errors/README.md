# Collecting errors from goroutines

A goroutine can't `return` an error to the code that started it. Yet when you
check 50 hosts at once, you need to know which ones failed, and why. This
exercise collects errors from concurrent work, reports all of them together,
and limits how many tasks run at once.

## Learn

### One error slot per task

The trick from *Goroutines: fix the maintenance script* works for errors too:
a slice with one slot per task, filled in by index:

```go
errs := make([]error, len(tasks))
var wg sync.WaitGroup
for i, task := range tasks {
	wg.Go(func() {
		errs[i] = task() // nil when the task succeeded
	})
}
wg.Wait()
```

### Joining them

As in *Reporting many errors*, `errors.Join` combines errors into one.
It **skips `nil`s**, and returns `nil` if every error is `nil`, so you can
pass the whole slice:

```go
return errors.Join(errs...)
```

The joined error's message is each message on its own line, in slice order
(so input order, not finish order). `errors.Is` still finds each one inside:

```text
web-2: connection refused
db-1: certificate expired
```

### Limiting concurrency with a semaphore

A *semaphore* is a counter of free slots. A buffered channel makes a good
one. Its capacity is the number of slots: sending takes a slot, and receiving
gives it back. When every slot is taken, the next send waits.

```go
sem := make(chan struct{}, 3) // at most 3 at once

for i, task := range tasks {
	sem <- struct{}{} // take a slot; waits while 3 tasks are running
	wg.Go(func() {
		defer func() { <-sem }() // give the slot back when done
		errs[i] = task()
	})
}
wg.Wait()
```

```text
limit 2, 4 tasks, each takes 1s:
slot 1:  task 0 ███  task 2 ███
slot 2:  task 1 ███  task 3 ███
         0s ─────── 1s ─────── 2s
```

`defer func() { <-sem }()` defers a call to an anonymous function, which
receives one value from `sem`. That frees a slot even if the task panics.

### Making tasks from data

A task here is a `func() error`. You can build one per host with a closure
that captures the loop variable, as in *Functions as values*:

```go
var tasks []func() error
for _, host := range hosts {
	tasks = append(tasks, func() error {
		return ping(host)
	})
}
```

## Your task

You run maintenance tasks across a fleet. Every task in the tests takes 1
(fake) second.

### 1. Run everything and report every error

`RunAll(tasks)` runs every task at the same time, waits for all of them, and
returns `errors.Join` of their errors, in task order. It returns `nil` if
every task succeeded.

```go
RunAll([]func() error{ok, failsWithDiskFull, ok, failsWithTimeout})
// => an error with message "disk full\ntimeout", after 1s
```

### 2. Run at most limit at once

`RunLimited(tasks, limit)` does the same, but never runs more than `limit`
tasks at once. Use a buffered channel as a semaphore.

```go
RunLimited(fiveTasks, 2) // never more than 2 running; takes 3s
```

### 3. Check hosts

`CheckHosts(hosts, limit, check)` calls `check(host)` for every host, with
at most `limit` at once, using `RunLimited`. Each failure is wrapped with
its host name as `"<host>: <error>"`, using `%w` so `errors.Is` still works.

```go
// check fails for web-2 with "connection refused"
CheckHosts([]string{"web-1", "web-2"}, 2, check)
// => an error with message "web-2: connection refused"
```

## Learn more

- [errors.Join](https://pkg.go.dev/errors#Join)
- [Go by Example: Errors](https://gobyexample.com/errors)
- [Effective Go: Channels as semaphores](https://go.dev/doc/effective_go#channels)
