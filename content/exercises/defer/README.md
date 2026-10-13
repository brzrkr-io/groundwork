# defer

Ops code is full of things that must be undone: a connection must be closed,
a lock must be released. `defer` makes sure the cleanup happens, even when the
function returns early because something failed.

## Learn

### What defer does

A `defer` statement schedules a function call to run **when the surrounding
function returns**, not right away:

```go
func Uptime(r *Recorder) {
	r.Connect("web-1")
	defer r.Close("web-1") // scheduled now, runs at the very end
	r.Run("uptime")
	r.Run("df")
}
// events: connect web-1, run uptime, run df, close web-1
```

### It runs on every return path

The deferred call runs no matter which `return` ends the function. That is the
whole point. Compare:

```go
// Leaks the connection when Run fails: the early return skips Close.
func Restart(r *Recorder) error {
	r.Connect("web-1")
	if err := r.Run("restart"); err != nil {
		return err
	}
	r.Close("web-1")
	return nil
}

// Always closes: defer runs on both returns.
func Restart(r *Recorder) error {
	r.Connect("web-1")
	defer r.Close("web-1")
	if err := r.Run("restart"); err != nil {
		return err
	}
	return nil
}
```

### Acquire, then defer straight away

Write the `defer` on the line right after you get the thing that needs
cleaning up. The cleanup sits next to the setup, so you can't forget it, and
every `return` below it is covered.

### Several defers run last-in, first-out

When a function defers more than one call, they run in reverse order: the last
one deferred runs first. That is exactly what cleanup needs, because you undo
things in the opposite order you did them:

```go
r.Lock("deploys")
defer r.Unlock("deploys") // runs second
r.Connect("web-1")
defer r.Close("web-1")    // runs first
// events: lock deploys, connect web-1, close web-1, unlock deploys
```

### Arguments are evaluated when you write defer

Go works out the arguments of a deferred call **at the `defer` line**, not when
the call finally runs:

```go
status := "failed"
defer r.Record("status " + status) // the text is built now: "status failed"
status = "ok"
// records "status failed"
```

If you want the value at the end, defer a *closure* (a `func() { ... }` that
you call with `()`). The closure reads `status` when it runs:

```go
defer func() {
	r.Record("status " + status)
}()
```

### Changing the returned error from a defer

A function can give its result a name: `func Backup(...) (err error)`. A
*named result* is a normal variable that holds what the function will return.
A deferred closure runs after `return` has set it, so the closure can still
change it. This adds context to every error in one place:

```go
func Restart(r *Recorder, host string) (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("restart %s: %w", host, err)
		}
	}()
	// every "return someErr" below gets the "restart <host>: " prefix
	...
}
```

## Your task

Your team's deploy tool talks to servers through a `Recorder`. It is a fake:
instead of contacting a real server, each method appends a line to
`r.Events`, so the tests can check what happened and in what order. The
`Recorder` is already written for you:

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

`Deploy` connects to the host, runs each command in order and stops at the
first command that fails, returning its error unchanged. Right now it only
closes the connection when every command succeeds. Use `defer` so the
connection is closed on every path.

```go
r := &Recorder{FailOn: "migrate"}
Deploy(r, "web-1", []string{"pull", "migrate", "restart"})
// => error "run migrate: command failed"
// r.Events: connect web-1, run pull, run migrate, close web-1
```

When every command works, `Deploy` returns `nil` and the log ends with
`close web-1` as before.

### 2. Lock, connect and clean up in order in Migrate

Write `Migrate(r, host)`. It must:

1. take the `migrations` lock and defer releasing it,
2. connect to `host` and defer closing it,
3. run `migrate` and return its error (or `nil`).

Thanks to last-in, first-out, the connection closes before the lock is
released, whether or not `migrate` fails:

```go
r := &Recorder{}
Migrate(r, "db-1")
// => nil
// r.Events: lock migrations, connect db-1, run migrate, close db-1, unlock migrations
```

### 3. Add context to every Backup error

`Backup(r, host)` connects, runs `dump`, then runs `upload`. Its result is
already named `err`. Add a deferred closure so that any error it returns is
wrapped with `%w` and prefixed with `backup <host>: `. Return `nil` unchanged
when everything works.

```go
r := &Recorder{FailOn: "upload"}
Backup(r, "db-1")
// => error "backup db-1: run upload: command failed"
// errors.Is(err, ErrCommandFailed) is true
// r.Events: connect db-1, run dump, run upload, close db-1
```

## Learn more

- [Go blog: Defer, Panic, and Recover](https://go.dev/blog/defer-panic-and-recover)
- [Go spec: Defer statements](https://go.dev/ref/spec#Defer_statements)
- [Go by Example: Defer](https://gobyexample.com/defer)
- [A Tour of Go: Defer](https://go.dev/tour/flowcontrol/12)
