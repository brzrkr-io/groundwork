# The race detector

A *data race* is a bug where two goroutines use the same variable at the same
time and at least one of them writes to it. Races give wrong totals, corrupt
maps and crash programs, often only under production load. Go's *race
detector* finds them for you. This exercise teaches you to read its reports.

**New here:** data races, `go test -race`, race reports. (`sync/atomic` is from *Mutexes: atomics and slow locks*.)

## Learn

### What counts as a data race

All three must be true:

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

`wg.Wait()` makes reading `count` **after** the loop safe. It does nothing
to stop the goroutines racing **with each other**.

> **Why is it written like this?** `count++` looks like one step, but the
> machine does three: read, add, write. Two goroutines can both read 5 and
> both write 6. Go requires you to make shared writes take turns; the
> compiler can't check it for you, which is why the detector exists.

### Turning on the detector

Add `-race` to `go test`, `go run` or `go build`:

```bash
go test -race ./...
```

Groundwork runs every test with `-race`. The detector only finds races in
code that actually runs, so test your concurrent code. Programs run slower
with `-race`: use it in tests, not production builds.

### Reading a report

A trimmed report for the code above, from `upload.go` in a project called
`backup`:

```text
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
...
--- FAIL: TestCountUploads (0.00s)
    testing.go:1867: race detected during execution of test
```

| Part | What it tells you |
|---|---|
| `Read at 0x… by goroutine 9` | One side: goroutine 9 **read** the variable at that address. |
| `CountUploads.func1()` … `upload.go:12` | Where: the anonymous function (`func1`) inside `CountUploads`, line 12. |
| `Previous write at` the **same address** | The other side: another goroutine **wrote** the same variable. |
| `Goroutine 9 (running) created at:` | Where that goroutine was started (`wg.Go`, line 11). |
| `race detected during execution of test` | The test fails because of the race, even if every value was right. |

How to use it:

1. Find **your** files in the stacks (skip `/usr/local/go/...`) and go to
   those lines.
2. Ask: which variable is used at both places? Here it is `count`.
3. Fix it, then run again: one report can hide another.

### Three ways to fix a race

| Fix | Use it for | Example |
|---|---|---|
| `sync/atomic` | one counter or flag | `var hits atomic.Int64` … `hits.Add(1)` … `hits.Load()` |
| `sync.Mutex` | maps, slices, several values that change together | `mu.Lock()`, `defer mu.Unlock()` |
| don't share | each goroutine writes its own slot, combine after `Wait` | `out[i] = ...` |

`atomic.Int64` methods take and return `int64`, so convert an `int` with
`int64(n)` and back with `int(x)`.

> **Why is it written like this?** Go requires a mutex to be shared, never
> copied, so you declare one `var mu sync.Mutex` next to the data it
> guards. Putting it right beside that data, with a `// protects results`
> comment, is convention. Which fix you pick is your choice.

### Maps can crash instead

A map written from several goroutines can crash the program with
`fatal error: concurrent map writes`. Same bug, the runtime just noticed it
first. Same fix: a mutex.

## Your task

A log-shipping agent measures files and checks hosts concurrently. Each
function has one data race. Fix each one so its test passes with no race
report. Keep the work concurrent: don't remove the goroutines.

### 1. Fix the byte counter

**Goal:** make `TotalBytes(files, size)` add up `size(file)` for every file
without a race.

Many goroutines update `total` at once. Use an `atomic.Int64` or a
`sync.Mutex`.

```go
TotalBytes([]string{"app.log", "error.log"}, size) // => 1500  (1200 + 300)
```

**Think it through:**
- Which variable do both sides of the race report point at?
- The function returns an `int`. What type does the atomic hold?

### 2. Fix the results map

**Goal:** make `CheckAll(hosts, check)` return a map from each host to
`check(host)` without a race.

Protect the map with a `sync.Mutex`. Run `check(host)` **before** taking the
lock, so slow checks still run at the same time.

```go
CheckAll([]string{"web-1", "web-2"}, check)
// => map[string]bool{"web-1": true, "web-2": false}
```

**Think it through:**
- Where does the mutex need to live so every goroutine shares the same one?
- Which part is slow (run it unlocked) and which part touches the map?

### 3. Fix the failed-hosts slice

**Goal:** make `FailedHosts(hosts, check)` return every host whose check
returned `false`, in any order, without a race.

`append` to a shared slice reads and writes it, so many goroutines appending
at once race. Protect it with a `sync.Mutex`.

```go
FailedHosts([]string{"web-1", "web-2", "db-1"}, check) // => []string{"web-2"}
```

**Think it through:**
- This is the same fix as task 2. Which single line needs the lock?

## Learn more

- [Data Race Detector](https://go.dev/doc/articles/race_detector)
- [Go blog: Introducing the Go Race Detector](https://go.dev/blog/race-detector)
- [The Go memory model](https://go.dev/ref/mem)
