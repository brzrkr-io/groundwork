## Task 1: one call does it

`errors.Join` skips `nil` arguments and returns `nil` when every argument is
`nil`. So you can pass the three check results straight in:

```go
return errors.Join(checkReachable(h), checkDisk(h), checkMemory(h))
```

## Task 2: collect, then join

Start with an empty `var errs []error` before the loops. Each element of
`checks` is a function, so `check(h)` runs it. Instead of returning the first
error, append it to `errs`. Append only the errors that are not `nil`.

## Task 2: adding the host name

Wrap each failure with `%w` so `errors.Is` still finds the sentinel:

```go
if err := check(h); err != nil {
	errs = append(errs, fmt.Errorf("%s: %w", h.Name, err))
}
```

After both loops, `return errors.Join(errs...)`. It is `nil` when `errs` is
empty.

## Task 3: order matters

Check for `nil` first, then for `ErrUnreachable`:

```go
if err == nil {
	return "deploy"
}
if errors.Is(err, ErrUnreachable) {
	return "abort"
}
return "warn"
```
