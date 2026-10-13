## Task 1: one keyword

Putting `go` in front of a function call starts it in a new goroutine and
moves on without waiting. Task 1 needs nothing else.

```go
go warm(host)
```

## Task 2: start, then wait

Create a `sync.WaitGroup` before the loop. Inside the loop, start each restart
with `wg.Go(func() { ... })`. After the loop, call `wg.Wait()`. Remember to
add `import "sync"` at the top of the file.

## Task 2: almost the answer

```go
var wg sync.WaitGroup
for _, service := range services {
	wg.Go(func() {
		restart(service)
	})
}
wg.Wait()
```

If the test says the restarts ran one after another, check that `wg.Wait()`
is **after** the loop, not inside it.
