# panic and recover

Most failures in Go are ordinary error values. A *panic* is different: it
stops the program. Ops tools such as job runners and servers need to know
when a panic is the right call, and how to stop one bad job from taking down
everything else.

**New here:** `panic`, `recover`, `Must...` functions.

## Learn

### What a panic is

A panic means "something is so wrong that this code cannot carry on". You
can cause one with the built-in `panic` function. Go also causes one when
your code breaks a rule while running.

```go
panic("config is missing") // you call panic yourself

var hosts []string
fmt.Println(hosts[3]) // index out of range: Go panics
```

`panic` accepts a value of any type. Usually it is a string or an error.

### What happens during a panic

1. The function stops at the panicking line.
2. Its deferred calls still run (as in *defer*).
3. The panic moves up to the caller, which stops too, and so on.
4. If nothing stops it, the program crashes and prints the panic value and a
   *stack trace* (the list of functions it was in).

```text
panic: config is missing

goroutine 1 [running]:
main.main()
	/app/main.go:5 +0x25
exit status 2
```

### When to panic, and when not to

| Situation | Use |
|---|---|
| Bad user input, network down, file missing | return an `error` (as in *Errors are values*) |
| A bug: "this can never happen" was wrong | panic (usually Go does it for you) |
| A value hard-coded in your program is invalid, checked at startup | a `Must...` function that panics |

The standard library follows the last row. `regexp.MustCompile` panics if the
pattern is broken, because the pattern is written in your source code: a
broken one is a bug you want to see the moment the program starts.

```go
var hostPattern = regexp.MustCompile(`^web-[0-9]+$`)
```

> **Why is it written like this?** The `Must` prefix is a convention, not a
> Go rule: it warns callers "this panics instead of returning an error".

### Stopping a panic with recover

The built-in `recover` stops a panic and returns the value passed to `panic`.

- It only works **directly inside a deferred function**. Anywhere else it does nothing and returns `nil`.
- It also returns `nil` when there is no panic.

Combine it with a named result to turn a panic into an error, like the
error-changing closures in *defer*:

```go
func parseConfig(path string) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("parsing %s crashed (%v)", path, r)
		}
	}()
	return riskyParse(path)
}
// riskyParse panics with "bad yaml"
// => error "parsing app.yaml crashed (bad yaml)"
```

> **Why is it written like this?** `if r := recover(); r != nil` declares `r`
> and checks it in one line, a common Go form. The name `r` is your choice.
> `r` has type `any`, so print it with `%v`, which works for every type.

### Keeping a worker alive

A job runner should not die because one job panicked. Run each job through a
helper that recovers, then carry on with the next job. A web server does the
same: one bad request gets an error page, everyone else is still served.

Only recover at a boundary like this, where you report the problem. Don't
sprinkle `recover` around to make bugs go quiet.

## Your task

You are writing the job runner for a nightly maintenance tool. It runs jobs
such as `backup` and `rotate` one after another and prints a report.

### 1. Panic on an invalid built-in port

**Goal:** make `MustPort(s)` return the port number, or panic with the string `invalid port "<s>"`.

`MustPort` is for ports hard-coded in the tool's source, so a bad one is a
bug. Parse `s` with `strconv.Atoi`. Panic when it is not a number, or when it
is outside `1`–`65535` (both ends allowed). Add the `fmt` and `strconv`
imports.

```go
MustPort("8080")  // => 8080
MustPort("abc")   // panics with the string: invalid port "abc"
MustPort("70000") // panics with the string: invalid port "70000"
MustPort("")      // panics with the string: invalid port ""
```

The panic value must be a **string** (not an error), and the double quotes
around `s` are part of the text.

**Think it through:**
- `strconv.Atoi` gives two results. Which one tells you "not a number"?
- There are three ways to be invalid. Can one `if` with `||` cover them all?
- Which `fmt` function builds a string, and which verb adds quotes around a value?

### 2. Turn a panic into an error

**Goal:** make `SafeRun(job)` recover a panic in `job` and return it as an error wrapping `ErrPanic`.

- The error text is `ErrPanic`'s text (`job panicked`), then `: `, then the panic value.
- Wrap `ErrPanic` with `%w` so `errors.Is(err, ErrPanic)` is true.
- If the job doesn't panic, return exactly what the job returned.

```go
SafeRun(func() error { return nil })
// => nil
SafeRun(func() error { return errors.New("disk full") })
// => the job's own error: "disk full"
SafeRun(func() error { panic("boom") })
// => error "job panicked: boom"
SafeRun(func() error { panic(42) })
// => error "job panicked: 42"
```

**Think it through:**
- `SafeRun`'s result is already named `err`. How does a deferred closure change it?
- Which verb wraps an error, and which prints a value of any type?
- When there is no panic, should your closure touch `err` at all?

### 3. Run every job, even after a panic

**Goal:** make `RunAll(jobs)` run each job with `SafeRun` and return one report line per job.

Each line is the job's `Name`, then `: `, then `ok` when the error is `nil`,
otherwise the error's text. A panicking job must not stop the jobs after it.
An empty list gives an empty report.

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

When Go itself panics inside a job, the panic value is Go's own message:
`vacuum: job panicked: runtime error: index out of range [2] with length 0`.

**Think it through:**
- Each `Job` has two fields. Which one do you pass to `SafeRun`, and which goes at the start of the line?
- An `error` isn't a string. Which method gives its text?
- Don't forget the space after the colon.

## Learn more

- [Go blog: Defer, Panic, and Recover](https://go.dev/blog/defer-panic-and-recover)
- [Effective Go: Panic](https://go.dev/doc/effective_go#panic) and [Recover](https://go.dev/doc/effective_go#recover)
- [Go by Example: Panic](https://gobyexample.com/panic) and [Recover](https://gobyexample.com/recover)
- [Go spec: Handling panics](https://go.dev/ref/spec#Handling_panics)
