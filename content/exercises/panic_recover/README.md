# panic and recover

Most failures in Go are ordinary error values. A *panic* is different: it
stops the program. Ops tools such as job runners and servers need to know
when a panic is the right call, and how to stop one bad job from taking down
everything else.

## Learn

### What a panic is

A panic means "something is so wrong that this code cannot carry on". You
cause one with the built-in `panic` function, and Go causes one for you when
your code breaks a rule at run time:

```go
panic("config is missing")   // you call panic yourself

var hosts []string
fmt.Println(hosts[3])          // index out of range: Go panics

var regions map[string]string
regions["web-1"] = "eu-west-1" // writing to a nil map: Go panics
```

When a panic starts, the function stops at that line. Deferred functions
still run (as in *defer*), then the panic moves up to the caller, which stops
too, and so on. If nothing stops it, the program crashes and prints the panic
value plus a *stack trace* (the list of functions it was in):

```text
panic: config is missing

goroutine 1 [running]:
main.main()
	/app/main.go:5 +0x25
exit status 2
```

`panic` accepts a value of any type. Usually it is a string or an error.

### When not to panic

A bad port in a config file, a host that doesn't answer, a missing file: those
are *expected* failures. Return an `error`, as in *Errors are values*, so the
caller can decide what to do.

| Situation | Use |
|---|---|
| Bad user input, network down, file missing | return an `error` |
| A bug: "this can never happen" was wrong | panic (usually Go does it for you) |
| A value hard-coded in your program is invalid, checked once at startup | a `Must...` function that panics |

The standard library follows the last row. `regexp.MustCompile` panics if the
pattern is broken, because the pattern is written in your source code: a
broken one is a bug you want to see the first time the program starts.

```go
var hostPattern = regexp.MustCompile(`^web-[0-9]+$`)
```

### Stopping a panic with recover

The built-in `recover` stops a panic. It only works when it is called
**directly inside a deferred function**. Anywhere else it does nothing and
returns `nil`. It also returns `nil` when there is no panic. During a panic it
returns the value that was passed to `panic`.

To turn a panic into an error, combine `recover` with a named result, just
like the error-annotating deferred functions in *defer*:

```go
func checkHost(name string) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("check %s panicked: %v", name, r)
		}
	}()
	return riskyCheck(name)
}
```

If `riskyCheck` panics with `"no route"`, the deferred function catches it,
sets `err`, and `checkHost` returns normally with the error
`check db-1 panicked: no route`. The type of `r` is `any`, so print it with
`%v`.

### Keeping a worker alive

A job runner that runs many jobs should not die because one of them panicked.
Run each job through a helper that recovers, then carry on with the next job.
This is what a web server does too: one bad request gets an error page and the
server keeps serving everyone else.

Recovering hides the crash, so only do it at a boundary like this, where you
report the problem. Don't sprinkle `recover` around to make bugs go quiet.

## Your task

You are writing the job runner for a nightly maintenance tool. It runs jobs
such as `backup` and `rotate` one after another and prints a report.

### 1. Panic on an invalid built-in port

`MustPort` is for ports hard-coded in the tool's source code, so a bad one is
a bug. Parse `s` with `strconv.Atoi`. If it is not a number, or it is outside
`1`–`65535`, panic with the string `invalid port "<s>"`. Otherwise return the
port. Add the `fmt` and `strconv` imports.

```go
MustPort("8080")  // => 8080
MustPort("abc")   // panics with the string: invalid port "abc"
MustPort("70000") // panics with the string: invalid port "70000"
```

### 2. Turn a panic into an error

`SafeRun` runs one job. Make it recover a panic. Then return an error that
wraps `ErrPanic` and includes the panic value:
`fmt.Errorf("%w: %v", ErrPanic, r)`. If the job does not panic, return
whatever the job returned.

```go
SafeRun(func() error { return nil })
// => nil
SafeRun(func() error { return errors.New("disk full") })
// => the job's own error: "disk full"
SafeRun(func() error { panic("boom") })
// => error "job panicked: boom", and errors.Is(err, ErrPanic) is true
```

### 3. Run every job, even after a panic

`RunAll` runs each job in order with `SafeRun` and returns one report line per
job: `<name>: ok` when the job's error is `nil`, otherwise `<name>: <error>`.
A panicking job must not stop the jobs after it.

```go
RunAll([]Job{
	{Name: "backup", Run: func() error { return nil }},
	{Name: "rotate", Run: func() error { panic("log dir missing") }},
	{Name: "prune", Run: func() error { return errors.New("no space left on device") }},
})
// => []string{
//     "backup: ok",
//     "rotate: job panicked: log dir missing",
//     "prune: no space left on device",
// }
```

An empty list of jobs gives an empty report. When Go itself panics inside a
job, the panic value is Go's own message, so the line reads like
`vacuum: job panicked: runtime error: index out of range [2] with length 0`.

## Learn more

- [Go blog: Defer, Panic, and Recover](https://go.dev/blog/defer-panic-and-recover)
- [Effective Go: Panic](https://go.dev/doc/effective_go#panic) and [Recover](https://go.dev/doc/effective_go#recover)
- [Go by Example: Panic](https://gobyexample.com/panic) and [Recover](https://gobyexample.com/recover)
- [Go spec: Handling panics](https://go.dev/ref/spec#Handling_panics)
