## Task 1: Run all checks on one host

Look at what `errors.Join` does with `nil` arguments. You don't need any `if`
statements here.

### Closer

`errors.Join` skips `nil`s and returns `nil` when all of them are `nil`, so
you can pass the three check results straight in, in order:
`errors.Join(checkReachable(h), ...)`.

### Answer

```go
return errors.Join(checkReachable(h), checkDisk(h), checkMemory(h))
```

Passing the checks in order puts their messages in that order, one per line.

## Task 2: Check every host

Keep the two loops, but instead of returning the first error, save it and
keep going. Join everything you saved at the end.

### Closer

Declare `var errs []error` before the loops. For each failure, append
`fmt.Errorf("%s: %w", h.Name, err)`. After the loops, spread the slice into
`errors.Join` with `errs...`.

### Answer

```go
var errs []error
for _, h := range hosts {
	for _, check := range checks {
		if err := check(h); err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", h.Name, err))
		}
	}
}
return errors.Join(errs...)
```

`errors.Join` of an empty slice is `nil`, and `%w` keeps each sentinel visible
to `errors.Is`.

## Task 3: Decide what to do

There are three cases, and the order you check them in matters. Handle `nil`
first, then look for `ErrUnreachable` anywhere in the error.

### Closer

`errors.Is(err, ErrUnreachable)` searches inside joined and wrapped errors.
A `switch { case err == nil: ... }` or a few `if`s both work.

### Answer

```go
switch {
case err == nil:
	return "deploy"
case errors.Is(err, ErrUnreachable):
	return "abort"
default:
	return "warn"
}
```
