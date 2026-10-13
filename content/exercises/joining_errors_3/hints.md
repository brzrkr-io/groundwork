## Task 1: Run every step

Loop over the steps and run each one. When a step fails, save the failure and
keep going instead of returning. Join everything after the loop.

### Closer

Declare `var errs []error`, then append `&StepError{Step: s.Name, Err: err}`
for each failure. Finish with `errors.Join(errs...)`, and add `"errors"` to
the import block.

### Answer

```go
var errs []error
for _, s := range steps {
	if err := s.Run(); err != nil {
		errs = append(errs, &StepError{Step: s.Name, Err: err})
	}
}
return errors.Join(errs...)
```

`errors.Join` of an empty slice is `nil`, so success needs no special case.

## Task 2: Split a joined error

Handle the three rows of the table in order. For the middle row, ask whether
`err` has an `Unwrap() []error` method with a type assertion.

### Closer

Check `err == nil` first. Then use
`joined, ok := err.(interface{ Unwrap() []error })`; when `ok`, return
`joined.Unwrap()`. Any other error becomes a one-element slice.

### Answer

```go
if err == nil {
	return nil
}
if joined, ok := err.(interface{ Unwrap() []error }); ok {
	return joined.Unwrap()
}
return []error{err}
```

## Task 3: List the failed steps

Use your `Split` to get the separate errors, then look for a `*StepError` in
each one. Collect the `Step` names in a slice.

### Closer

Range over `Split(err)`. For each error `e`, use
`stepErr, ok := errors.AsType[*StepError](e)` and append `stepErr.Step` when
`ok` is true. Start with `var names []string` so a `nil` error gives `nil`.

### Answer

```go
var names []string
for _, e := range Split(err) {
	if stepErr, ok := errors.AsType[*StepError](e); ok {
		names = append(names, stepErr.Step)
	}
}
return names
```

`errors.AsType` also finds a `*StepError` that was wrapped with `%w`.
