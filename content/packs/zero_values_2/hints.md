## Task 1: Fix the log path default

Read the `if` line out loud. `path != ""` means "path is NOT empty". When
should the default `"/var/log/app.log"` be used: when `path` is empty, or
when it isn't?

### Closer

Use the default only when `path` is empty, so the comparison needs `==`
instead of `!=`: `if path == "" {`. Everything else stays the same.

### Answer

```go
if path == "" {
	return "/var/log/app.log"
}
return path
```

## Task 2: Default the sample rate

This is the same pattern as task 1, but the parameter is a `float64`. An
unset rate arrives as `0`.

### Closer

Check `rate == 0` with an `if`. Inside it, return `0.1`. After the `if`,
return `rate` unchanged.

### Answer

```go
if rate == 0 {
	return 0.1
}
return rate
```

## Task 3: Finish the rotation summary

Look at the `keepFiles` check and read the variable on the left of `=`. Is it
the variable that was checked? Then notice there's no check for
`maxSizeMiB` at all.

### Closer

Inside `if keepFiles == 0 {`, assign to `keepFiles`, not `maxSizeMiB`. Then
add a second `if` of the same shape that sets `maxSizeMiB` to `100` when it
is `0`. The `fmt.Sprintf` line is already correct.

### Answer

```go
if maxSizeMiB == 0 {
	maxSizeMiB = 100
}
if keepFiles == 0 {
	keepFiles = 5
}
```
