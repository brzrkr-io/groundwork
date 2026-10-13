# defer: fix the log rotation job

`defer` is simple to write and easy to get subtly wrong. This exercise has
three classic `defer` bugs from a real-looking log rotation job. Find them
and fix them.

## Learn

Recap from the previous exercise: a deferred call runs when the surrounding
returns, on every return path, and several defers run last-in, first-out. The
arguments of a deferred call are evaluated at the `defer` line. A deferred
closure reads variables when it runs, at the end.

Two more rules matter here:

- **Check the error before you defer the cleanup.** If acquiring failed,
  there is nothing to clean up. Defer only after the `if err != nil` check:

  ```go
  if err := r.Connect(host); err != nil {
  	return err
  }
  defer r.Close(host)
  ```

- **A defer in a loop waits for the function, not the loop.** Deferred calls
  run when the *function* returns, so a `defer` inside a `for` loop keeps
  every resource open until the whole loop is over. The fix is to move the
  loop body into its own function. Its defers then run at the end of each
  call, which is the end of each iteration:

  ```go
  for _, host := range hosts {
  	if err := handleOne(r, host); err != nil { // handleOne has the defer
  		return err
  	}
  }
  ```

## Your task

The nightly job connects to each host and rotates its logs. As in the
previous exercise, the `Recorder` is a fake that appends a line to `r.Events`
for every call:

| Call | Appends to `r.Events` | Fails when |
|---|---|---|
| `r.Connect("web-1")` | `connect web-1` | host is `r.Unreachable`: `connect web-1: connection refused` |
| `r.Rotate("web-1")` | `rotate web-1` | host is `r.DiskFull`: `rotate web-1: disk full` |
| `r.Close("web-1")` | `close web-1` | never |
| `r.Record("any text")` | `any text` | never |

`Connect` and `Rotate` append their event even when they fail. Every function
below returns the failing call's error unchanged.

### 1. Record the real outcome in Maintenance

`Maintenance(r, host)` connects, rotates and closes, then records a final
line `maintenance <host>: ok` when everything worked or
`maintenance <host>: failed` when anything failed. Right now it always
records `failed`. Fix it so the deferred record sees the final `status`.

```go
r := &Recorder{}
Maintenance(r, "web-1")
// => nil
// r.Events: connect web-1, rotate web-1, close web-1, maintenance web-1: ok

r = &Recorder{Unreachable: "web-3"}
Maintenance(r, "web-3")
// => error "connect web-3: connection refused"
// r.Events: connect web-3, maintenance web-3: failed
```

### 2. Don't close what you never opened in RotateHost

`RotateHost(r, host)` connects, rotates and closes. When `Connect` fails it
still calls `Close`, which closes a connection that never existed. Fix it so a
failed connect records only `connect <host>`:

```go
r := &Recorder{Unreachable: "web-3"}
RotateHost(r, "web-3")
// => error "connect web-3: connection refused"
// r.Events: connect web-3

r = &Recorder{DiskFull: "web-2"}
RotateHost(r, "web-2")
// => error "rotate web-2: disk full"
// r.Events: connect web-2, rotate web-2, close web-2
```

### 3. Close each host before the next in RotateAll

`RotateAll(r, hosts)` rotates every host in order and stops at the first
error. Because its `defer` sits inside the loop, every connection stays open
until the end. Fix it so each host is closed before the next one is
contacted. The easiest fix is to call your fixed `RotateHost` from the loop,
so finish task 2 first.

```go
r := &Recorder{}
RotateAll(r, []string{"web-1", "web-2"})
// => nil
// r.Events: connect web-1, rotate web-1, close web-1, connect web-2, rotate web-2, close web-2

r = &Recorder{DiskFull: "web-2"}
RotateAll(r, []string{"web-1", "web-2", "web-3"})
// => error "rotate web-2: disk full"
// r.Events: connect web-1, rotate web-1, close web-1, connect web-2, rotate web-2, close web-2
```

## Learn more

- [Go spec: Defer statements](https://go.dev/ref/spec#Defer_statements)
- [Effective Go: Defer](https://go.dev/doc/effective_go#defer)
- [Go blog: Defer, Panic, and Recover](https://go.dev/blog/defer-panic-and-recover)
