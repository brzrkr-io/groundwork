## Task 1: range over the channel

`for line := range lines { ... }` receives one line per loop and stops when
the channel is closed. Append each line to `collected`.

## Task 2: a mailbox big enough for everything

`make(chan string, len(alerts))` has room for every alert, so sending them
never waits. Send each alert with `ch <- alert`, then `close(ch)`, then
`return ch`.

## Task 3: send from a goroutine

An unbuffered send waits for a receiver, and nobody is receiving until
`StreamLines` returns. So the sending has to happen in a goroutine:

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
