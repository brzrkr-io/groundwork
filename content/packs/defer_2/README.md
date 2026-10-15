# defer: fix the log rotation job

`defer` is simple to write and easy to get subtly wrong. This exercise has
three classic `defer` bugs from a real-looking log rotation job. Find them
and fix them.

**New here:** deferring after an error check, `defer` inside loops. Otherwise practice.

## Learn

### Recap from *defer*

- A deferred call runs when the surrounding **function** returns, on every return path.
- Several defers run last-in, first-out.
- A deferred call's **arguments** are worked out at the `defer` line.
- A deferred **closure** (`defer func() { ... }()`) reads variables at the end.

### Check the error before you defer the cleanup

If acquiring something failed, there is nothing to clean up. So the order is:
acquire, check the error, **then** defer the cleanup.

```go
f, err := os.Open("app.log")
if err != nil {
	return err // nothing was opened, nothing to close
}
defer f.Close() // only reached when Open worked
```

> **Why is it written like this?** Go doesn't force this order; the compiler
> accepts a `defer` anywhere. Open, check, defer is the Go convention, and
> skipping it causes bugs like the one in task 2.

### A defer in a loop waits for the function, not the loop

Deferred calls run when the *function* returns. A `defer` inside a `for` loop
therefore piles up: nothing is cleaned until the whole loop is over.

```go
for _, name := range files {
	f, _ := os.Open(name)
	defer f.Close() // every file stays open until the function ends
}
```

The fix: move one iteration's work into its own function. That function's
defers run each time it returns, which is the end of each iteration.

```go
for _, name := range files {
	if err := printFile(name); err != nil { // printFile opens, defers Close, reads
		return err
	}
}
```

> **Why is it written like this?** Moving the loop body into a helper is the
> common Go idiom; other fixes exist, but this one is the clearest.

## Your task

The nightly job connects to each host and rotates its logs. As in *defer*,
the `Recorder` is a fake that appends a line to `r.Events` for every call:

| Call | Appends to `r.Events` | Fails when |
|---|---|---|
| `r.Connect("web-1")` | `connect web-1` | host is `r.Unreachable`: `connect web-1: connection refused` |
| `r.Rotate("web-1")` | `rotate web-1` | host is `r.DiskFull`: `rotate web-1: disk full` |
| `r.Close("web-1")` | `close web-1` | never |
| `r.Record("any text")` | `any text` | never |

`Connect` and `Rotate` append their event even when they fail. Every function
below returns the failing call's error unchanged.

### 1. Record the real outcome in Maintenance

**Goal:** make `Maintenance` record `maintenance <host>: ok` or `maintenance <host>: failed`, matching what really happened.

`Maintenance(r, host)` connects, rotates and closes, then records a final
line. Right now it always records `failed`, even on success.

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

**Think it through:**
- When is the text `"maintenance " + host + ": " + status` built: at the `defer` line or at the end?
- What value does `status` have at that moment?
- Which form of `defer` (from the recap) reads `status` at the end instead?

### 2. Don't close what you never opened in RotateHost

**Goal:** make `RotateHost` skip `Close` when `Connect` failed.

`RotateHost(r, host)` connects, rotates and closes. A failed connect must
record only `connect <host>`.

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

**Think it through:**
- Look at the order of the three lines: connect, defer, check. Which should come before the `defer`?
- After a failed connect, the function returns. Has the `defer` been scheduled yet?

### 3. Close each host before the next in RotateAll

**Goal:** make `RotateAll` close each host before connecting to the next one.

`RotateAll(r, hosts)` rotates every host in order and stops at the first
error. Because its `defer` sits inside the loop, every connection stays open
until the end. Finish task 2 first: your fixed `RotateHost` already does one
host correctly.

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

**Think it through:**
- Which function already does "connect, rotate, close" for one host?
- What should the loop do with the error that function returns?

## Learn more

- [Go spec: Defer statements](https://go.dev/ref/spec#Defer_statements)
- [Effective Go: Defer](https://go.dev/doc/effective_go#defer)
- [Go blog: Defer, Panic, and Recover](https://go.dev/blog/defer-panic-and-recover)
