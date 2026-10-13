# The race detector

A *data race* is a bug where two goroutines use the same variable at the same
time and at least one of them writes to it. Races give wrong totals, corrupt
maps and crash programs, often only under production load. Go's *race
detector* finds them for you. This exercise teaches you to read its reports.

## Learn

### What counts as a data race

All three of these must be true:

1. Two or more goroutines access the same variable.
2. At least one access is a write.
3. Nothing makes them take turns: no mutex, atomic, channel or `wg.Wait()`
   in between.

```go
count := 0
var wg sync.WaitGroup
for range buckets {
	wg.Go(func() {
		count++ // read count, add 1, write count: in many goroutines at once
	})
}
wg.Wait()
```

The `wg.Wait()` makes reading `count` **after** the loop safe. It does
nothing to stop the goroutines racing **with each other**.

### Turning on the detector

Add `-race` to `go test`, `go run` or `go build`:

```bash
go test -race ./...
go run -race .
```

Gopherlings runs every test in this module with `-race`. The detector watches
memory accesses as the program runs, so it only finds races in code that
actually runs. That's one more reason to test your concurrent code. Programs
run slower with `-race`, so it is for tests and staging, not production
builds.

### Reading a report

Here is a trimmed report for the code above, from a file `upload.go` in a
project called `backup`:

```text
==================
WARNING: DATA RACE
Read at 0x00c000126168 by goroutine 9:
  example.com/backup.CountUploads.func1()
      /home/you/backup/upload.go:12 +0x67

Previous write at 0x00c000126168 by goroutine 10:
  example.com/backup.CountUploads.func1()
      /home/you/backup/upload.go:12 +0x7e

Goroutine 9 (running) created at:
  sync.(*WaitGroup).Go()
      /usr/local/go/src/sync/waitgroup.go:238 +0x72
  example.com/backup.CountUploads()
      /home/you/backup/upload.go:11 +0x88

Goroutine 10 (finished) created at:
  sync.(*WaitGroup).Go()
      /usr/local/go/src/sync/waitgroup.go:238 +0x72
  example.com/backup.CountUploads()
      /home/you/backup/upload.go:11 +0x88
==================
--- FAIL: TestCountUploads (0.00s)
    testing.go:1867: race detected during execution of test
```

| Part | What it tells you |
|---|---|
| `Read at 0x00c000126168 by goroutine 9` | One side of the race: goroutine 9 **read** the variable at that memory address. |
| `CountUploads.func1()` … `upload.go:12` | Where: in the anonymous function (`func1`) inside `CountUploads`, line 12. |
| `Previous write at` the **same address** `by goroutine 10` | The other side: goroutine 10 **wrote** the same variable, at `upload.go:12` too. |
| `Goroutine 9 (running) created at:` | Where each goroutine was started: `wg.Go`, called from line 11. |
| `race detected during execution of test` | The test failed because of the race, even if every value it checked was right. |

How to use it:

1. Find **your** files in the stacks (skip lines from `/usr/local/go/...` or
   the Go installation) and go to those line numbers.
2. Ask: which variable is used at both places? Here it is `count`.
3. Pick a fix. Then run again, because one race report can hide another.

### Three ways to fix a race

| Fix | Use it for | Example |
|---|---|---|
| `sync/atomic` | a single counter or flag | `var count atomic.Int64` … `count.Add(1)` |
| `sync.Mutex` | maps, slices, several values that change together | `mu.Lock()`, `defer mu.Unlock()` |
| don't share | each goroutine writes its own slot, combine after `Wait` | `results[i] = ...` |

### Maps can crash instead

A map written from several goroutines can also crash the whole program with
`fatal error: concurrent map writes`. That is the same bug, Go's runtime just
noticed it first. The fix is the same: a mutex.

## Your task

A log-shipping agent measures files and checks hosts concurrently. Each
function below has one data race. Fix each one so its test passes with no
race report. Keep the work concurrent: don't remove the goroutines.

### 1. Fix the byte counter

`TotalBytes(files, size)` adds up `size(file)` for every file. Many
goroutines update `total` at once. Make the sum safe, using an
`atomic.Int64` or a `sync.Mutex`.

```go
TotalBytes([]string{"app.log", "error.log"}, size) // => 1500  (1200 + 300)
```

### 2. Fix the results map

`CheckAll(hosts, check)` returns a map from each host to `check(host)`. Many
goroutines write the map at once. Protect it with a `sync.Mutex`. Run
`check(host)` **before** taking the lock, so that slow checks still run at the
same time.

```go
CheckAll([]string{"web-1", "web-2"}, check)
// => map[string]bool{"web-1": true, "web-2": false}
```

### 3. Fix the failed-hosts slice

`FailedHosts(hosts, check)` returns every host whose check returned `false`,
in any order. `append` to a shared slice from many goroutines is a race,
because `append` reads and writes the slice. Protect it with a `sync.Mutex`.

```go
FailedHosts([]string{"web-1", "web-2", "db-1"}, check) // => []string{"web-2"}
```

## Learn more

- [Data Race Detector](https://go.dev/doc/articles/race_detector)
- [Go blog: Introducing the Go Race Detector](https://go.dev/blog/race-detector)
- [The Go memory model](https://go.dev/ref/mem)
