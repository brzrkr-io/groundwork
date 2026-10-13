## Task 1: Collect lines from a channel

You need to keep receiving until the sender closes the channel. One kind of loop does exactly that.

### Closer

`for line := range lines { ... }` receives one value per loop and stops once the channel is closed and empty. Inside the loop, use `append` to add to `collected`.

### Answer

```go
for line := range lines {
	collected = append(collected, line)
}
return collected
```

## Task 2: Queue alerts in a buffered channel

A buffered channel with room for every alert lets you send them all without anyone receiving. Fill it, then close it so a `range` over it ends.

### Closer

Make it with `make(chan string, len(alerts))`. Send each alert with `ch <- alert`, then call `close(ch)` and return `ch`.

### Answer

```go
ch := make(chan string, len(alerts))
for _, alert := range alerts {
	ch <- alert
}
close(ch)
return ch
```

The buffer has one slot per alert, so no send ever waits.

## Task 3: Stream lines through an unbuffered channel

A send on an unbuffered channel waits for a receiver, and nobody receives until `StreamLines` has returned. So who should do the sending?

### Closer

Make the channel with `make(chan string)`, do the sending and the `close` inside `go func() { ... }()`, and `return ch` after starting the goroutine.

### Answer

```go
ch := make(chan string)
go func() {
	for _, line := range lines {
		ch <- line
	}
	close(ch)
}()
return ch
```

The goroutine sends each line as the caller receives it, then closes the channel so the caller's loop ends.
