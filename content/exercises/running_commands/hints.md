## Task 1: keep what Output returns

`cmd.Output()` returns two values: the stdout bytes and an error. Convert the
bytes with `string(out)` and trim them with `strings.TrimSpace`.

```go
out, err := cmd.Output()
return strings.TrimSpace(string(out)), err
```

## Task 2: pull the exit code out of the error

Check for `*exec.ExitError` first. Only if it isn't one, and `err` is still
not `nil`, did the program fail to start.

```go
if exitErr, ok := errors.AsType[*exec.ExitError](err); ok {
	return exitErr.ExitCode(), strings.TrimSpace(string(out)), nil
}
if err != nil {
	return -1, "", err
}
```

Add `"errors"` to the imports.

## Task 3: wire up the options

```go
ctx, cancel := context.WithTimeout(context.Background(), opts.Timeout)
defer cancel()
cmd := exec.CommandContext(ctx, name, args...)
cmd.Dir = opts.Dir
cmd.Env = append(os.Environ(), opts.Env...)
```

## Task 3: explain the timeout

Inside `if err != nil`, ask the context whether its deadline passed:

```go
if ctx.Err() != nil {
	return "", fmt.Errorf("timed out after %v: %w", opts.Timeout, ctx.Err())
}
return "", err
```
