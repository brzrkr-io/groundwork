# Running commands: output, exit codes and timeouts

Ops tools constantly run other tools: `systemctl`, `kubectl`, `git`,
`terraform`. Go's `os/exec` package starts a program, collects what it prints
and tells you how it exited, with no shell in between. And with the context
you met in *Context and cancellation*, a hung tool can't hang your tool too.

**New here:** `os/exec`, `*exec.ExitError`, `exec.CommandContext`.

## Learn

### Starting a program

```go
cmd := exec.Command("git", "rev-parse", "HEAD")
out, err := cmd.Output()
```

`exec.Command(name, args...)` only *describes* the program. Nothing runs
until you call a method:

| Method | Runs the program and returns |
|---|---|
| `cmd.Run()` | only the error |
| `cmd.Output()` | stdout as `[]byte`, and the error |
| `cmd.CombinedOutput()` | stdout and stderr mixed, and the error |

Output is a `[]byte` that usually ends in `"\n"`. `string(out)` turns it into
a string; `strings.TrimSpace` removes the spaces and newlines around it.

> **Why is it written like this?** Go starts the program *directly*, with no
> shell. Each argument arrives exactly as written, so `*` is never expanded
> and `"web-1; rm -rf /"` stays one harmless argument. That's why arguments
> are separate strings, never one joined command line.

### Exit codes

A program reports success with its *exit code*, as in *Exit codes, && and
||* and *Standard streams: stdin, stdout, stderr and exit codes*: `0` means
success, anything else is failure. `systemctl is-active nginx` prints
`inactive` and exits with `3` when nginx is stopped.

A non-zero exit gives an error of type `*exec.ExitError`. Get it with
`errors.AsType`, as in *Wrapping errors*:

```go
if e, ok := errors.AsType[*exec.ExitError](err); ok {
	fmt.Println("exit code:", e.ExitCode()) // 3
}
```

If the program couldn't start at all (no such file), the error is a
different type, and there's no exit code.

### Directory and environment

```go
cmd.Dir = "/srv/app"                                 // run in this directory
cmd.Env = append(os.Environ(), "LANG=C", "TZ=UTC")   // ours plus two more
```

`os.Environ()` is our own environment as `"KEY=value"` strings. Leave
`cmd.Env` as `nil` and the program inherits everything. Set it **without**
`os.Environ()` and the program loses `PATH`, `HOME` and the rest.

> **Why is it written like this?** `append(list, a, b)` returns a new list,
> so you must store the result (Go requires using it). To append a whole
> slice, put `...` after it: `append(list, more...)`.

### Timeouts with a context

In *Context and cancellation* you built contexts with `context.WithTimeout`
and watched `ctx.Done()` yourself. `exec.CommandContext` does that watching
for you: it starts the program and, if the context ends first, kills it.

```go
ctx, stop := context.WithTimeout(context.Background(), 30*time.Second)
defer stop()
cmd := exec.CommandContext(ctx, "terraform", "plan")
```

When the deadline passes, Go kills the program and the error just says
`signal: killed`. Afterwards `ctx.Err()` is `context.DeadlineExceeded`, which
tells you *why*. Wrap it with `%w` so callers can check it with `errors.Is`.

> **Why is it written like this?** As in *Context and cancellation*, always
> call the cancel function (`go vet` warns if you don't), and `defer` is the
> convention. The name (`cancel`, `stop`) is your choice. The context is the
> first argument, as in every context-aware Go API.

### How the tests fake a tool

The tests can't rely on real tools, so the test binary pretends to be one.
`TestMain` in `exercise_test.go` checks for `RUNCMD_HELPER=1`; if set, it
runs `fakeToolMain` instead of tests. The fake commands are `echo`, `exit`,
`pwd`, `getenv` and `sleep`. The child only sees `RUNCMD_HELPER=1` if it
inherits our environment.

## Your task

You're writing the helpers a deploy tool uses to run other programs. In the
examples, `tool` is the fake tool (`os.Args[0]`).

### 1. Capture stdout

**Goal:** make `Output` return the program's stdout, trimmed, plus the error.

Switch from `cmd.Run()` to `cmd.Output()`. Stderr is not included. A
non-zero exit returns an error.

```go
Output(tool, "echo", "  nginx is running  ")
// => "nginx is running", nil
```

**Think it through:**
- `Output()` returns two things. Which goes back as the string?
- What turns `[]byte` into a string, and what trims it?

### 2. Report the exit code

**Goal:** make `ExitStatus` return the exit code and trimmed combined output,
with a `nil` error whenever the program ran.

It already runs `CombinedOutput`. Three cases:

| What happened | Return |
|---|---|
| ran, exit `0` | `0`, trimmed output, `nil` |
| ran, non-zero exit | that exit code, trimmed output, `nil` |
| couldn't start | `-1`, `""`, the error |

```go
ExitStatus(tool, "exit", "3", "inactive", "")
// => 3, "inactive", nil
ExitStatus("/no/such/tool", "status")
// => -1, "", <error>
```

Add `"errors"` to the imports.

**Think it through:**
- Which error type means "it ran but failed"? Check for that one first.
- After that check, what does any other non-`nil` error mean?

### 3. Directory, environment and timeout

**Goal:** make `Run` use `opts.Dir`, `opts.Env` and `opts.Timeout`, and
return trimmed stdout.

- Make a context that times out after `opts.Timeout`; build the command with
  `exec.CommandContext`.
- `cmd.Dir` is `opts.Dir`.
- `cmd.Env` is our environment plus every entry of `opts.Env`.
- If `cmd.Output()` fails **and** `ctx.Err()` isn't `nil`, return `""` and
  `fmt.Errorf("timed out after %v: %w", opts.Timeout, ctx.Err())`.

Add `"context"`, `"os"` and `"fmt"` to the imports.

```go
Run(Options{Env: []string{"DEPLOY_ENV=staging"}, Timeout: 5 * time.Second}, tool, "getenv", "DEPLOY_ENV")
// => "staging", nil
Run(Options{Timeout: 50 * time.Millisecond}, tool, "sleep")
// => "", timed out after 50ms: context deadline exceeded
```

The fake `sleep` lasts 2 seconds. If the timeout test takes that long, the
timeout isn't wired up yet.

**Think it through:**
- Which line in `Run` creates `cmd`? That's the one to replace.
- `opts.Env` is already a slice. How do you append a whole slice?

## Learn more

- [os/exec package](https://pkg.go.dev/os/exec)
- [Go by Example: Spawning Processes](https://gobyexample.com/spawning-processes)
- [context.WithTimeout](https://pkg.go.dev/context#WithTimeout)
