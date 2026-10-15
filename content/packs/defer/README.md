# defer

Ops code is full of things that must be undone: a connection must be closed,
a lock must be released. `defer` makes sure the cleanup happens, even when the
function returns early because something failed.

**New here:** `defer`, last-in first-out order, deferred closures, changing named results (from *Multiple return values: don't drop the error*) in a deferred function.

## Learn

### What defer does

A `defer` statement schedules a function call to run **when the surrounding
function returns**, not right away.

```go
func Report(log *Log) {
	log.Open("report.txt")
	defer log.Save("report.txt") // scheduled now, runs at the very end
	log.Write("disk 80%")
	log.Write("cpu 12%")
}
// order: open, write, write, save
```

> **Why is it written like this?** Go requires `defer` to be followed by a
> function *call* (with `()`), not just a function name. Putting the `defer`
> on the line right after the setup is convention: the cleanup sits next to
> the thing it cleans up.

### It runs on every return path

The deferred call runs no matter which `return` ends the function.

```go
func Check(log *Log) error {
	log.Open("check.txt")
	defer log.Save("check.txt")
	if err := log.Write("ping"); err != nil {
		return err // Save still runs here
	}
	return nil // ...and here
}
```

Without `defer`, a cleanup line placed at the bottom is **skipped** by every
early `return` above it.

### Several defers run last-in, first-out

When a function defers more than one call, the last one deferred runs first.
That is what cleanup needs: you undo things in the opposite order you did
them, like taking off shoes before socks.

```go
defer fmt.Println("first deferred")  // prints second
defer fmt.Println("second deferred") // prints first
```

### Arguments are worked out at the defer line

Go evaluates the arguments of a deferred call **when the `defer` line runs**,
not when the call finally happens.

```go
status := "failed"
defer fmt.Println("status " + status) // text built now: "status failed"
status = "ok"
// prints "status failed"
```

To read a variable's value at the end, defer a *closure*: a
`func() { ... }` that you call right away with `()`. Its body runs at the end,
so it sees the latest value.

```go
defer func() {
	fmt.Println("status " + status) // prints "status ok"
}()
```

> **Why is it written like this?** The trailing `()` is required: `defer`
> needs a call, and `func() {...}` alone is only a value.

### Named results: changing what a function returns

A function can name its result: `func Fetch() (err error)`. A *named result*
is a normal variable that holds what the function will return. `return x`
copies `x` into it, and **then** deferred calls run, so a deferred closure can
still change it.

```go
func Count() (n int) {
	defer func() {
		n = n * 10 // runs after "return 4" has set n to 4
	}()
	return 4
}
// Count() => 40
```

> **Why is it written like this?** The parentheses around `(n int)` are
> required once a result has a name. The name itself (`n`, `err`) is your
> choice; `err` is the convention for errors.

## Your task

Your team's deploy tool talks to servers through a `Recorder`. It is a fake:
each method appends a line to `r.Events` instead of contacting a server, so
the tests can check what happened and in what order. It is already written:

| Call | Appends to `r.Events` |
|---|---|
| `r.Connect("web-1")` | `connect web-1` |
| `r.Close("web-1")` | `close web-1` |
| `r.Lock("migrations")` | `lock migrations` |
| `r.Unlock("migrations")` | `unlock migrations` |
| `r.Run("migrate")` | `run migrate` |

`Run` fails for the command named in `r.FailOn`. It still appends its event,
then returns an error wrapping `ErrCommandFailed`, for example
`run migrate: command failed`.

### 1. Always close the connection in Deploy

**Goal:** make `Deploy` close the connection on every return path, using `defer`.

`Deploy` connects, runs each command in order, and stops at the first failing
command, returning its error unchanged. Right now it only closes when every
command succeeds.

```go
r := &Recorder{FailOn: "migrate"}
Deploy(r, "web-1", []string{"pull", "migrate", "restart"})
// => error "run migrate: command failed"
// r.Events: connect web-1, run pull, run migrate, close web-1
```

When every command works (or there are none), `Deploy` returns `nil` and the
events still end with `close <host>`.

**Think it through:**
- Which line does the setup, and which line is the cleanup?
- Where should the cleanup move to, so no `return` can skip it? Does the old line still belong at the bottom?
- Use the `host` parameter, not a fixed `"web-1"`.

### 2. Lock, connect and clean up in order in Migrate

**Goal:** write `Migrate(r, host)` so it locks, connects, runs `migrate`, and cleans up in reverse order.

It must:

1. take the lock named `migrations` and defer releasing it,
2. connect to `host` and defer closing it,
3. run `migrate` and return its error (or `nil`).

```go
r := &Recorder{}
Migrate(r, "db-1")
// => nil
// r.Events: lock migrations, connect db-1, run migrate, close db-1, unlock migrations
```

The events are the same when `migrate` fails; only the returned error is
`run migrate: command failed`.

**Think it through:**
- `"migrations"` is fixed text; the host comes from the parameter.
- With last-in, first-out, which `defer` must you write first so `unlock` comes last?
- `Run` already returns an error. Can you return its result directly?

### 3. Add context to every Backup error

**Goal:** add a deferred closure to `Backup` that prefixes any returned error with `backup <host>: `.

`Backup(r, host)` connects, runs `dump`, then runs `upload`. Its result is
already named `err`. Wrap with `%w` so `errors.Is(err, ErrCommandFailed)`
stays true. When everything works, return `nil` unchanged.

```go
r := &Recorder{FailOn: "upload"}
Backup(r, "db-1")
// => error "backup db-1: run upload: command failed"
// r.Events: connect db-1, run dump, run upload, close db-1
```

**Think it through:**
- Inside the closure, check the named result: is there an error to wrap at all?
- You set the new error by assigning to `err`, not by `return` (the closure returns nothing).
- The pieces: fixed text `backup `, the `host` parameter, fixed `: `, then the old error with `%w` (as in *Errors are values*).

## Learn more

- [Go blog: Defer, Panic, and Recover](https://go.dev/blog/defer-panic-and-recover)
- [Go spec: Defer statements](https://go.dev/ref/spec#Defer_statements)
- [Go by Example: Defer](https://gobyexample.com/defer)
- [A Tour of Go: Defer](https://go.dev/tour/flowcontrol/12)
