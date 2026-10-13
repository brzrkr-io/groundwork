## Task 1: Fix the deadlock

Each `answers <- ...` waits for a receiver, but the receiving loop only starts after `wg.Wait()`. How could the sends finish without anyone receiving yet?

### Closer

Give the channel a buffer with one slot per host. `make(chan string, n)` makes a buffered channel; a send only waits when the buffer is full.

### Answer

```go
answers := make(chan string, len(hosts))
```

With one slot per host, every send finishes straight away, so `wg.Wait()` can return.

## Task 2: Prefix lines with direction types

Receive from `in` until it's closed, and send a changed copy of each line to `out`. Keep the `close(out)` the starter already has, after your loop.

### Closer

Use `for line := range in { ... }` and send with `out <- ...`. Build the text with `host + ": " + line`.

### Answer

```go
for line := range in {
	out <- host + ": " + line
}
close(out)
```

## Task 3: Take a few jobs

Loop at most `n` times and receive one job each time. The comma-ok receive tells you when the channel has closed, so you can stop early.

### Closer

Use `for range n { ... }` and receive with `job, ok := <-jobs`. If `ok` is false, return `taken, true`; after the loop, return `taken, false`.

### Answer

```go
for range n {
	job, ok := <-jobs
	if !ok {
		return taken, true
	}
	taken = append(taken, job)
}
return taken, false
```

`ok` is only false when the channel is closed and empty, so getting all `n` jobs always returns `false`.
