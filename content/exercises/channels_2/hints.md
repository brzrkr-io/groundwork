## Task 1: room in the buffer

With an unbuffered channel, every `answers <- ...` waits for a receiver, but
the receiving loop only starts after `wg.Wait()`. A buffered channel with one
slot per host lets every send finish straight away:
`make(chan string, len(hosts))`.

## Task 2: range, then send

```go
for line := range in {
	out <- host + ": " + line
}
```

Keep the `close(out)` after the loop.

## Task 3: count up to n

Loop `n` times. Each time, receive with `job, ok := <-jobs`. If `ok` is
false, the channel is closed: `return taken, true`. Otherwise append the job.
After the loop, `return taken, false`.

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
