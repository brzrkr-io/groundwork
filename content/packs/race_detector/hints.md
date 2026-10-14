## Task 1: Fix the byte counter

Read the `WARNING: DATA RACE` report: both sides point at the `total +=` line.
A single number that many goroutines add to is the job `sync/atomic` was made
for.

### Closer

Declare the counter as `var total atomic.Int64` (import `"sync/atomic"`).
Goroutines call `total.Add(...)`, which takes an `int64`, and you read the sum
with `total.Load()` after `wg.Wait()`.

### Answer

```go
var total atomic.Int64
var wg sync.WaitGroup
for _, file := range files {
	wg.Go(func() {
		total.Add(int64(size(file)))
	})
}
wg.Wait()
return int(total.Load())
```

`Add` does the read, add and write as one step, so no update is lost.

## Task 2: Fix the results map

Many goroutines write `results` at once. A map needs a `sync.Mutex`, declared
next to it, and only the write should happen while holding the lock.

### Closer

Run the slow `check(host)` first and keep its result in a variable. Then
take the lock just for the map write:

```go
ok := check(host)
mu.Lock()
defer mu.Unlock()
// write the map here
```

### Answer

```go
results := make(map[string]bool)
var mu sync.Mutex // protects results
var wg sync.WaitGroup
for _, host := range hosts {
	wg.Go(func() {
		ok := check(host) // the slow part runs without the lock
		mu.Lock()
		defer mu.Unlock()
		results[host] = ok
	})
}
wg.Wait()
return results
```

Checks still run at the same time; only the quick map write takes turns.

## Task 3: Fix the failed-hosts slice

`append` reads and writes the slice header, so many goroutines appending to
`failed` race. Protect it the same way as the map in task 2.

### Closer

Add `var mu sync.Mutex` next to `failed`, and wrap only the `append` with
`mu.Lock()` and `defer mu.Unlock()`, inside the `if !check(host)` block.

### Answer

```go
var failed []string
var mu sync.Mutex // protects failed
var wg sync.WaitGroup
for _, host := range hosts {
	wg.Go(func() {
		if !check(host) {
			mu.Lock()
			defer mu.Unlock()
			failed = append(failed, host)
		}
	})
}
wg.Wait()
return failed
```

The lock makes each `append` finish before the next one starts.
