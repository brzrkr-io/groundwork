## Read the report first

Scroll up from the failing test to the `WARNING: DATA RACE` block. The first
lines that mention `exercise.go` show the line number of the racing access.
Each TODO marks one of them.

## Task 1: an atomic counter

Replace `total := 0` with `var total atomic.Int64` (import `"sync/atomic"`).
Inside the goroutine: `total.Add(int64(size(file)))`. At the end:
`return int(total.Load())`.

## Tasks 2 and 3: a mutex next to the data

Declare `var mu sync.Mutex` next to `results` (or `failed`). In the goroutine,
do the slow `check` first, then lock just around the shared write:

```go
ok := check(host)
mu.Lock()
defer mu.Unlock()
results[host] = ok
```

For task 3, lock around `failed = append(failed, host)` the same way.
