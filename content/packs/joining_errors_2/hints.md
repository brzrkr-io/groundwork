## Task 1: Report every problem with a rule

`ValidateRule` already has a `var errs []error` and joins it at the end, but
nothing ever lands in it. Each check returns right away instead of saving its
problem.

### Closer

Turn each `return fmt.Errorf(...)` into `errs = append(errs, fmt.Errorf(...))`.
Keep the messages exactly as they are, and keep `return errors.Join(errs...)`
as the last line.

### Answer

```go
if r.Port < 1 || r.Port > 65535 {
	errs = append(errs, fmt.Errorf("rule %q: %w: %d", r.Name, ErrBadPort, r.Port))
}
if r.Protocol != "tcp" && r.Protocol != "udp" {
	errs = append(errs, fmt.Errorf("rule %q: %w: %q", r.Name, ErrBadProtocol, r.Protocol))
}
if r.Source == "0.0.0.0/0" && r.Port != 80 && r.Port != 443 {
	errs = append(errs, fmt.Errorf("rule %q: %w: port %d", r.Name, ErrOpenToWorld, r.Port))
}
return errors.Join(errs...)
```

## Task 2: Return nil when every rule is fine

`fmt.Errorf` with `%w` always returns a real error, even when you wrap `nil`.
Find out whether there is anything to report before you wrap it.

### Closer

Store `errors.Join(errs...)` in a variable first. Only wrap it with
`"check firewall: %w"` when it is not `nil`; otherwise return `nil`.

### Answer

```go
if err := errors.Join(errs...); err != nil {
	return fmt.Errorf("check firewall: %w", err)
}
return nil
```

`errors.Join` already skips the `nil` results from good rules, so it is `nil`
only when no rule had a problem.

## Task 3: Find open rules inside a report

`==` only matches the bare sentinel. Use the `errors` function that searches
inside wrapped and joined errors.

### Closer

`errors.Is(err, target)` looks through `%w` and `errors.Join`, and returns
`false` for `nil`.

### Answer

```go
return errors.Is(err, ErrOpenToWorld)
```
