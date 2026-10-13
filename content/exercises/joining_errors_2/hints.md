## Task 1: append instead of return

`ValidateRule` already has a `var errs []error` and joins it at the end, but
nothing ever lands in it. Each `return fmt.Errorf(...)` should become an
`append` to `errs`.

## Task 1: almost the answer

```go
if r.Port < 1 || r.Port > 65535 {
	errs = append(errs, fmt.Errorf("rule %q: %w: %d", r.Name, ErrBadPort, r.Port))
}
```

Do the same for the other two checks, and leave `return errors.Join(errs...)`
as the last line.

## Task 2: join first, wrap second

Store the joined error in a variable and only wrap it when it is not `nil`:

```go
if err := errors.Join(errs...); err != nil {
	return fmt.Errorf("check firewall: %w", err)
}
return nil
```

## Task 3: errors.Is

`errors.Is(err, ErrOpenToWorld)` looks through wrapping and joining, and
returns `false` for a `nil` error.
