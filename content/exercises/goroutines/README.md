# Goroutines

A *goroutine* is a function running at the same time as the rest of your
program. Ops tools use them constantly: restart ten services at once instead of
one after another, or check fifty hosts in the time it takes to check one.

## Learn

### One thing at a time

Normally Go runs one statement after another. If each restart takes 1 second,
three restarts take 3 seconds:

```go
restart("api")    // 1s
restart("web")    // 1s
restart("worker") // 1s
// 3 seconds in total
```

### The go keyword

Put `go` in front of a function call and Go starts that call in a new
goroutine, then carries on **immediately** without waiting for it:

```go
go restart("api")
go restart("web")
go restart("worker")
// we get here straight away; the restarts are still running
```

```text
main      ──go──go──go──▶ carries on at once
api         └─ restart("api") ──────▶
web            └─ restart("web") ───▶
worker            └─ restart("worker") ▶
```

You can also start an anonymous function, which is handy when you need a few
lines of work. Note the `()` at the end: it calls the function.

```go
go func() {
	fmt.Println("restarting", name)
	restart(name)
}()
```

### Nobody waits for a goroutine

`go` gives you nothing back: no return value and no way to know when the work
is done. When `main` returns, the program exits, and any goroutines still
running are simply stopped:

```go
func main() {
	go fmt.Println("hello from a goroutine")
}
// usually prints nothing: main returned before the goroutine ran
```

Sometimes "start it and move on" is exactly what you want (*fire and
forget*). Usually you need to wait for the work to finish.

### Waiting with sync.WaitGroup

A `sync.WaitGroup` from the `sync` package counts running goroutines and lets
you wait until they have all finished. Since Go 1.25, `wg.Go` starts a
goroutine and counts it in one step:

```go
import "sync"

var wg sync.WaitGroup
for _, name := range services {
	wg.Go(func() {
		restart(name)
	})
}
wg.Wait() // blocks until every function started with wg.Go has returned
```

```text
main   ──Go──Go──Go──Wait ··········· (blocked) ··········▶ carries on
api          └─ restart("api") ──────┘
web              └─ restart("web") ───┘
worker               └─ restart("worker") ┘
```

The zero value of `sync.WaitGroup` is ready to use, so `var wg
sync.WaitGroup` is all the setup you need. You'll meet the rest of the
WaitGroup API in *WaitGroups*.

### Loop variables

The function passed to `wg.Go` uses `name` from the loop. Since Go 1.22, every
loop iteration gets its own copy of `name`, so each goroutine sees the right
service. (Older Go versions shared one variable across iterations, and you
will still see code that works around that with `name := name`.)

### About the tests

The tests use `testing/synctest`, which gives them a fake clock. A
`time.Sleep(time.Second)` inside the test finishes instantly in real time, but
the test can still measure that one (fake) second passed. That is how the tests
tell "all at once" (1s) from "one after another" (3s).

## Your task

Your deploy tool restarts services and warms caches one host at a time. Make it
concurrent.

### 1. Warm caches in the background

`WarmCaches(hosts, warm)` should call `warm(host)` for every host, each in its
own goroutine, and return **straight away** without waiting for them. Nothing
calls `WarmCaches` back; this is fire and forget.

```go
WarmCaches([]string{"cache-1", "cache-2"}, warm)
// returns immediately; warm("cache-1") and warm("cache-2") run in the background
```

### 2. Restart services and wait

`RestartAll(services, restart)` should call `restart(service)` for every
service at the same time and return only when **all** of them have finished.
Use a `sync.WaitGroup` and `wg.Go`. You need to add `"sync"` to the imports.

If each restart takes 1 second, restarting three services should take about 1
second, not 3.

```go
RestartAll([]string{"api", "web", "worker"}, restart)
// returns after ~1s, once all three restarts are done
```

## Learn more

- [A Tour of Go: Goroutines](https://go.dev/tour/concurrency/1)
- [Go by Example: Goroutines](https://gobyexample.com/goroutines)
- [Go by Example: WaitGroups](https://gobyexample.com/waitgroups)
- [sync.WaitGroup](https://pkg.go.dev/sync#WaitGroup)
