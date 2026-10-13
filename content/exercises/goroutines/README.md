# Goroutines

A *goroutine* is a function running at the same time as the rest of your
program. Ops tools use them to restart ten services at once, or check fifty
hosts in the time it takes to check one.

**New here:** the `go` keyword, `sync.WaitGroup`, `wg.Go`, `wg.Wait`.

## Learn

### One thing at a time

Normally Go runs one statement after another. If each backup takes 1 second,
three backups take 3 seconds:

```go
backup("users")  // 1s
backup("orders") // 1s
backup("logs")   // 1s
// 3 seconds in total
```

### The go keyword

Put `go` in front of a function **call**. Go starts that call in a new
goroutine and carries on **immediately**, without waiting:

```go
go backup("users")
go backup("orders")
// we get here straight away; both backups are still running
```

```text
main    ──go──go──▶ carries on at once
users     └─ backup("users") ──────▶
orders       └─ backup("orders") ──▶
```

You can also start an anonymous function. The `()` at the end calls it:

```go
go func() {
	fmt.Println("backing up", table)
	backup(table)
}()
```

> **Why is it written like this?** Go requires `go` to be followed by a
> function *call* (with its `()` and arguments), not just a function name.
> `go backup` without `("users")` does not compile.

### Nobody waits for a goroutine

`go` gives you nothing back: no return value, and no way to know when the
work is done. When `main` returns, the program exits and any running
goroutines are stopped:

```go
func main() {
	go fmt.Println("hello from a goroutine")
}
// usually prints nothing: main returned first
```

"Start it and move on" is called *fire and forget*. Often you need to wait
instead.

### Waiting with sync.WaitGroup

A `sync.WaitGroup` counts running goroutines and lets you wait for all of
them. `wg.Go(f)` starts `f` in a goroutine and counts it. `wg.Wait()` blocks
until every counted function has returned:

```go
import "sync"

var wg sync.WaitGroup
for _, table := range tables {
	wg.Go(func() {
		backup(table)
	})
}
wg.Wait() // waits for every backup
```

```text
main    ──Go──Go──Wait ····· (blocked) ·····▶ carries on
users       └─ backup("users") ──────┘
orders         └─ backup("orders") ──┘
```

> **Why is it written like this?**
> - Go requires `import "sync"` before you can write `sync.WaitGroup`.
> - `var wg sync.WaitGroup` is all the setup: Go guarantees its zero value
>   is ready to use.
> - The name `wg` is a convention; any name works.
> - `wg.Wait()` goes **after** the loop (author's choice that matters):
>   inside the loop it would wait for each goroutine before starting the
>   next.

You'll meet more of the WaitGroup API in *WaitGroups*.

### Loop variables

The function passed to `wg.Go` uses `table` from the loop. Since Go 1.22,
each loop iteration gets its own copy, so each goroutine sees the right
value. Older code works around this with `table := table`; you don't need to.

### About the tests

The tests use `testing/synctest`, which gives them a fake clock. A one-second
sleep finishes instantly in real time, but the test can still see that one
(fake) second passed. That is how it tells "all at once" (1s) from "one after
another" (3s).

## Your task

Your deploy tool restarts services and warms caches one host at a time. Make
it concurrent.

### 1. Warm caches in the background

**Goal:** make `WarmCaches(hosts, warm)` start `warm(host)` for every host in
its own goroutine and return straight away.

Don't wait for the warm-ups. This is fire and forget.

```go
WarmCaches([]string{"cache-1", "cache-2"}, warm)
// returns immediately; both warm-ups run in the background
```

**Think it through:**
- Find the line inside the loop that calls `warm`. That call is what should
  run in its own goroutine.
- Which one keyword turns a call into "start it and move on"?

### 2. Restart services and wait

**Goal:** make `RestartAll(services, restart)` run every `restart(service)`
at the same time and return only when **all** of them have finished.

Use a `sync.WaitGroup` with `wg.Go`, and add `"sync"` to the imports. If each
restart takes 1 second, three services should take about 1 second, not 3.

```go
RestartAll([]string{"api", "web", "worker"}, restart)
// returns after ~1s, once all three restarts are done
```

**Think it through:**
- Where does the WaitGroup need to be declared so every loop iteration uses
  the same one: before, inside or after the loop?
- `wg.Go` takes a function, not a call. What wraps `restart(service)`?
- Where does `wg.Wait()` go so the restarts still overlap?

## Learn more

- [A Tour of Go: Goroutines](https://go.dev/tour/concurrency/1)
- [Go by Example: Goroutines](https://gobyexample.com/goroutines)
- [Go by Example: WaitGroups](https://gobyexample.com/waitgroups)
- [sync.WaitGroup](https://pkg.go.dev/sync#WaitGroup)
