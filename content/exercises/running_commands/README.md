# Running commands: output, exit codes and timeouts

Ops tools constantly run other tools: `systemctl`, `kubectl`, `git`,
`terraform`. Go's `os/exec` package starts a program, collects what it prints
and tells you how it exited, all without a shell in between.

## Learn

### Starting a program

```go
cmd := exec.Command("systemctl", "is-active", "nginx")
out, err := cmd.Output()
```

`exec.Command(name, args...)` describes the program and its arguments. Nothing
runs until you call a method such as `Output`. Go starts the program
*directly*, without a shell. Each argument reaches the program exactly as you
wrote it, so `*` is never expanded and a hostname like `web-1; rm -rf /` stays
one harmless argument. Never build a command by joining strings.

| Method | Runs the program and returns |
|---|---|
| `cmd.Run()` | only the error |
| `cmd.Output()` | stdout as `[]byte`, and the error |
| `cmd.CombinedOutput()` | stdout and stderr mixed together, and the error |

The output is a `[]byte` that usually ends in `"\n"`, so convert it with
`string(out)` and tidy it with `strings.TrimSpace`.

### Exit codes and errors

A program reports success or failure with its *exit code*: `0` means success,
anything else is a failure. Scripts rely on this. `systemctl is-active nginx`
prints `inactive` and exits with code `3` when nginx is stopped.

When a program exits with a non-zero code, the error is an `*exec.ExitError`.
Use `errors.AsType` (from *Wrapping errors*) to get it:

```go
if exitErr, ok := errors.AsType[*exec.ExitError](err); ok {
	fmt.Println("exit code:", exitErr.ExitCode()) // 3
}
```

With `Output`, whatever the program wrote to stderr ends up in
`exitErr.Stderr`. If the program could not be started at all, for example
because no such file exists, the error is a different type, and there is no
exit code to report.

### Directory, environment and timeouts

```go
cmd.Dir = "/srv/app"                                      // run in this directory
cmd.Env = append(os.Environ(), "DEPLOY_ENV=staging")      // our variables plus one more
```

`os.Environ()` returns our own environment as `"KEY=value"` strings. If you
leave `cmd.Env` as `nil`, the program inherits it all. If you set `cmd.Env`
**without** starting from `os.Environ()`, the program loses `PATH`, `HOME` and
everything else, and tools break in confusing ways.

To stop a program that hangs, use this recipe. A *context* carries a deadline.
You'll learn contexts properly in the Concurrency module.

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()
cmd := exec.CommandContext(ctx, "terraform", "plan")
out, err := cmd.Output()
if ctx.Err() != nil {
	// the deadline passed and the program was killed
	return fmt.Errorf("timed out after %v: %w", 5*time.Second, ctx.Err())
}
```

When the deadline passes, Go kills the program and `err` just says
`signal: killed`. `ctx.Err()` returns `context.DeadlineExceeded`, which tells
you *why*. Wrap it with `%w` so callers can check it with `errors.Is`.

### How the tests fake a tool

The tests can't rely on `systemctl` being installed, so they use the *helper
process* pattern, like the tests of `os/exec` itself. The test file has a
`TestMain` function that runs before any test. If the environment variable
`RUNCMD_HELPER=1` is set, the test binary behaves like a small fake tool
instead of running tests. Each test sets that variable and runs `os.Args[0]`,
which is the test binary itself. Look at `fakeToolMain` in
`exercise_test.go` to see the fake commands (`echo`, `exit`, `pwd`, `getenv`,
`sleep`). The child only sees `RUNCMD_HELPER=1` because it inherits our
environment, which is one more reason to start `cmd.Env` from `os.Environ()`.

## Your task

You're writing the helpers a deploy tool uses to run other programs. In the
examples, `tool` is the fake tool from the tests (`os.Args[0]`).

### 1. Capture stdout

Fix `Output` so it runs the program with `cmd.Output()` and returns stdout
with surrounding whitespace trimmed, plus the error. Stderr is not included.
A non-zero exit returns an error.

```go
Output(tool, "echo", "  nginx is running  ")
// => "nginx is running", nil
```

### 2. Report the exit code

Finish `ExitStatus`. It already runs the program with `CombinedOutput`.

- If the program ran (whatever its exit code), return the exit code, the
  trimmed combined output and a `nil` error. A non-zero exit is not an error
  here; the caller reads the code.
- If the program could not be started, return `-1`, `""` and the error.

```go
ExitStatus(tool, "exit", "3", "inactive", "")
// => 3, "inactive", nil
ExitStatus("/no/such/tool", "status")
// => -1, "", <error>
```

### 3. Directory, environment and timeout

Finish `Run`, which returns trimmed stdout like `Output`.

- Use `context.WithTimeout` with `opts.Timeout` and `exec.CommandContext`.
- Set `cmd.Dir` to `opts.Dir`.
- Set `cmd.Env` to `os.Environ()` plus `opts.Env`.
- If `cmd.Output()` fails and `ctx.Err()` is not `nil`, return
  `fmt.Errorf("timed out after %v: %w", opts.Timeout, ctx.Err())`. The tests
  check it with `errors.Is(err, context.DeadlineExceeded)`.

```go
Run(Options{Env: []string{"DEPLOY_ENV=staging"}, Timeout: 5 * time.Second}, tool, "getenv", "DEPLOY_ENV")
// => "staging", nil
Run(Options{Timeout: 50 * time.Millisecond}, tool, "sleep")
// => "", timed out after 50ms: context deadline exceeded
```

The fake `sleep` lasts 2 seconds. If the timeout test takes that long and
fails, the timeout isn't wired up yet.

## Learn more

- [os/exec package](https://pkg.go.dev/os/exec)
- [Go by Example: Spawning Processes](https://gobyexample.com/spawning-processes)
- [context.WithTimeout](https://pkg.go.dev/context#WithTimeout)
