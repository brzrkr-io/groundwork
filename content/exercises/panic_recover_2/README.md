# panic and recover: fix the chat bot's safety net

A teammate tried to make the team's chat bot crash-proof. The idea was right,
but `recover` is easy to use wrongly. In this exercise you fix three common
mistakes.

## Learn

Recap: a panic stops the current function and climbs up through its callers,
running deferred functions on the way. `recover()` stops it, but only when it
is called directly inside a deferred function. A deferred function can change
a named result, which is how a panic becomes an error. Recover at a boundary,
like a job runner or a chat-command handler, and never to hide a bug.

Two details matter for the bugs below.

**`recover` outside a deferred function does nothing.** Code before the panic
runs before anything has gone wrong, so `recover()` returns `nil` there. Code
after the panic never runs at all. Only deferred functions run *during* the
panic.

**After a panic, the `return` statement never runs.** The function's results
are whatever they hold when the deferred functions finish. With an unnamed
result, that is the zero value, and a deferred function has no way to change
it:

```go
func check() error {
	var err error
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("check panicked: %v", r) // sets a local variable...
		}
	}()
	err = probe() // panics
	return err    // ...but this line never runs, so check returns nil
}
```

With `func check() (err error)`, the deferred function sets the result itself,
so the caller gets the error.

## Your task

The team's chat bot runs commands such as `restart api` and keeps a summary of
firing alerts. Right now one bad command can take the whole bot offline, and
the alert summary is always empty.

### 1. Make Guard stop the panic

`Guard` runs the handler for one chat command. It calls `recover()` before the
handler even runs, so a panicking handler still crashes the bot. Move the
`recover` into a deferred function so that a panic no longer escapes from
`Guard`.

When the handler doesn't panic, `Guard` returns whatever the handler returned
(`nil` or its error). That part already works. Keep it working.

### 2. Return the panic as an error

After task 1 the panic is stopped, but `Guard` returns `nil`, so the bot
thinks the command succeeded. Make `err` a named result so the deferred
function can set it. The error message is:

```text
command "<command>" panicked: <panic value>
```

```go
Guard("restart api", func() error { panic("boom") })
// => error `command "restart api" panicked: boom`

Guard("scale web", func() error { panic(7) })
// => error `command "scale web" panicked: 7`
```

### 3. Fix the real bug instead of recovering

`CountBySeverity` counts alerts per severity. It panics on every alert, and
someone "fixed" that by recovering and ignoring the panic. Now it quietly
returns `nil`.

Writing to a `nil` map always panics. That is a bug in the code, not a
problem to recover from. Delete the deferred `recover` and fix the bug: create
the map with `make` before the loop. With no alerts, return an empty map, not
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

## Learn more

- [Go blog: Defer, Panic, and Recover](https://go.dev/blog/defer-panic-and-recover)
- [Go spec: Handling panics](https://go.dev/ref/spec#Handling_panics)
- [Go blog: Go maps in action](https://go.dev/blog/maps)
