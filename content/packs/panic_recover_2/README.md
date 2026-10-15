# panic and recover: fix the chat bot's safety net

A teammate tried to make the team's chat bot crash-proof. The idea was right,
but `recover` is easy to use wrongly. In this exercise you fix three common
mistakes.

**New here:** why `recover` must be deferred, why the result must be named. Otherwise practice.

## Learn

### Recap from *panic and recover*

- A panic stops the current function and climbs up through its callers, running deferred calls on the way.
- `recover()` stops it, but only when called **directly inside a deferred function**.
- A deferred closure can change a **named result**: that is how a panic becomes an error.
- Recover at a boundary (a job runner, a chat-command handler), never to hide a bug.

### recover outside a deferred function does nothing

Only deferred calls run *during* a panic. So anywhere else, `recover` is too
early or too late:

| Where `recover()` is | What happens |
|---|---|
| before the panicking line | nothing has gone wrong yet: returns `nil` |
| after the panicking line | never runs at all |
| inside a deferred function | runs during the panic and stops it |

### After a panic, the return line never runs

When a panic is recovered, the function ends with its results as they are
**after the deferred calls finish**. The `return` line itself was skipped.

```go
func load() int {
	n := 0
	defer func() {
		recover()
		n = -1 // changes the local n...
	}()
	n = readCount() // panics
	return n        // ...but this never runs, so load returns 0
}
```

With an unnamed result, the caller gets the zero value and the deferred
function can't change that. With a named result, `func load() (n int)`, the
deferred function sets the result itself and the caller gets `-1`.

> **Why is it written like this?** Go requires parentheses around named
> results: `(n int)`. A named result is already declared, so writing
> `n := 0` or `var n int` again in the body is a compile error. Delete the old
> declaration line; don't edit the line the error points at (see *Reading
> compiler errors*).

### Writing to a nil map panics

`var m map[string]int` declares a map but creates nothing: `m` is `nil`.
Reading from it is fine; writing to it panics. `make` creates a real, empty
map you can write to.

```go
var seen map[string]bool // nil: seen["a"] = true would panic
ports := make(map[string]int) // ready: ports["ssh"] = 22 works
```

## Your task

The team's chat bot runs commands such as `restart api` and keeps a summary of
firing alerts. Right now one bad command can take the whole bot offline, and
the alert summary is always empty.

### 1. Make Guard stop the panic

**Goal:** make `Guard` stop a panic in `handler`, so no panic escapes from `Guard`.

`Guard` calls `recover()` before the handler even runs, so a panicking
handler still crashes the bot. When the handler doesn't panic, `Guard` must
still return whatever the handler returned (`nil` or its error, unchanged).

**Think it through:**
- Look at the table above: where is the `recover` now, and where must it be?
- Which form of `defer` lets you run several lines (the `if` and the assignment)?

### 2. Return the panic as an error

**Goal:** make `Guard` return an error describing the panic instead of `nil`.

After task 1 the panic is stopped, but `Guard` returns `nil`, so the bot
thinks the command worked. The error text is:

```text
command "<command>" panicked: <panic value>
```

The quotes around the command are part of the text.

```go
Guard("restart api", func() error { panic("boom") })
// => error `command "restart api" panicked: boom`

Guard("scale web", func() error { panic(7) })
// => error `command "scale web" panicked: 7`
```

**Think it through:**
- Which line declares `err` now? Where must the name `err` live instead?
- Once it is named, which old line becomes a duplicate declaration?
- The message format is already in the code. Does it still need to change?

### 3. Fix the real bug instead of recovering

**Goal:** make `CountBySeverity` return the real count per severity, without any `recover`.

It panics on every alert, and someone "fixed" that by recovering and
ignoring the panic, so it quietly returns `nil`. Delete the deferred
`recover` and fix the actual bug. With no alerts, return an empty map, not
`nil`.

```go
CountBySeverity([]Alert{
	{Name: "DiskFull", Severity: "critical"},
	{Name: "HighLatency", Severity: "warning"},
	{Name: "PodCrashLooping", Severity: "critical"},
})
// => map[string]int{"critical": 2, "warning": 1}

CountBySeverity(nil)
// => map[string]int{}
```

**Think it through:**
- Which line declares `counts`? Is that map `nil` or real?
- The loop line that panics is fine; the fix belongs on the declaration.

## Learn more

- [Go blog: Defer, Panic, and Recover](https://go.dev/blog/defer-panic-and-recover)
- [Go spec: Handling panics](https://go.dev/ref/spec#Handling_panics)
- [Go blog: Go maps in action](https://go.dev/blog/maps)
