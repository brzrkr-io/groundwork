## Task 1: two one-liners

`Error` builds a string from `e.Field`, `": "` and `e.Err.Error()`.
`Unwrap` just returns `e.Err`. That one line is what lets `errors.Is` see the
category inside a `*FieldError`.

## Tasks 2 and 3: one if per rule

Each rule is an `if` that appends to `errs`:

```go
if c.Port < 1 || c.Port > 65535 {
	errs = append(errs, &FieldError{
		Field: "port",
		Err:   fmt.Errorf("%w: %d is not between 1 and 65535", ErrOutOfRange, c.Port),
	})
}
```

Keep the order from the README: name, image, port, replicas, env. For the
image, check `c.Image == ""` first, then use `else if` with
`_, tag, found := strings.Cut(c.Image, ":")`.

## Task 3: checking an env name

```go
for _, r := range name {
	ok := r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '_'
	if !ok {
		return false
	}
}
```

Return `false` for an empty name before the loop. In `Validate`, loop over
`slices.Sorted(maps.Keys(c.Env))`.

## Tasks 4 and 5: a shared helper

```go
func fieldErrors(err error) []*FieldError {
	if err == nil {
		return nil
	}
	list := []error{err}
	if joined, ok := err.(interface{ Unwrap() []error }); ok {
		list = joined.Unwrap()
	}
	var found []*FieldError
	for _, e := range list {
		if fe, ok := errors.AsType[*FieldError](e); ok {
			found = append(found, fe)
		}
	}
	return found
}
```

`FieldNames` loops over `fieldErrors(err)` and collects `fe.Field`.
`FieldsWith` does the same, but only when `errors.Is(fe, kind)` is true.
