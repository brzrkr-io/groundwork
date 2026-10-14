## Task 1: Capture stdout

`cmd.Output()` runs the program and gives back what it wrote to stdout, plus
an error. Keep both, and turn the bytes into a trimmed string.

### Closer

Convert with `string(out)` and trim with `strings.TrimSpace`:

```go
out, err := cmd.Output()
```

### Answer

```go
cmd := exec.Command(name, args...)
out, err := cmd.Output()
return strings.TrimSpace(string(out)), err
```

## Task 2: Report the exit code

A program that ran but exited non-zero gives an `*exec.ExitError`. Check for
that type first; any other non-`nil` error means the program never started.

### Closer

`errors.AsType[*exec.ExitError](err)` returns the typed error and an `ok`.
The exit code is `exitErr.ExitCode()`. Add `"errors"` to the imports.

### Answer

```go
out, err := exec.Command(name, args...).CombinedOutput()
if exitErr, ok := errors.AsType[*exec.ExitError](err); ok {
	return exitErr.ExitCode(), strings.TrimSpace(string(out)), nil
}
if err != nil {
	return -1, "", err
}
return 0, strings.TrimSpace(string(out)), nil
```

## Task 3: Directory, environment and timeout

Make a context with a deadline, build the command from it, and set the `Dir`
and `Env` fields before running. When it fails, ask the context whether the
deadline was the reason.

### Closer

```go
ctx, cancel := context.WithTimeout(context.Background(), opts.Timeout)
defer cancel()
cmd := exec.CommandContext(ctx, name, args...)
```

`cmd.Env` should be `append(os.Environ(), opts.Env...)`. After a failed
`cmd.Output()`, `ctx.Err() != nil` means the timeout killed it.

### Answer

```go
ctx, cancel := context.WithTimeout(context.Background(), opts.Timeout)
defer cancel()
cmd := exec.CommandContext(ctx, name, args...)
cmd.Dir = opts.Dir
cmd.Env = append(os.Environ(), opts.Env...)

out, err := cmd.Output()
if err != nil {
	if ctx.Err() != nil {
		return "", fmt.Errorf("timed out after %v: %w", opts.Timeout, ctx.Err())
	}
	return "", err
}
return strings.TrimSpace(string(out)), nil
```
