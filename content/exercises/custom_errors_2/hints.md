## Task 1: Fix the error message

The format string only has a verb for the port. It needs one more for the owner, and the wording must match the README.

### Closer

Use `%d` for the port and `%q` for the owner (it adds the double quotes), and pass `e.Port` and `e.Owner` in that order.

### Answer

```go
return fmt.Sprintf("port %d is already used by %q", e.Port, e.Owner)
```

## Task 2: Fix the typed-nil bug in Reserve

On the free-port path, `err` is a nil `*PortInUseError`. Returning it as an `error` fills in the type part of the interface, so the result is not equal to `nil`.

### Closer

Remove the `var err *PortInUseError` line. Return the `&PortInUseError{...}` straight from inside the `if`, and end the function with a literal `return nil`.

### Answer

```go
if owner, ok := used[port]; ok {
	return &PortInUseError{Port: port, Owner: owner}
}
used[port] = service
return nil
```

A literal `nil` returned as `error` has no type and no value, so `err == nil` is true for the caller.

## Task 3: Finish Advice

You need to ask whether `err` holds a `*PortInUseError` and, if so, get at its `Owner` and `Port` fields. Do it before the `"unexpected error: "` line.

### Closer

Use the two-result type assertion, which never panics: `perr, ok := err.(*PortInUseError)`. When `ok` is true, `perr` is the concrete error.

### Answer

```go
if perr, ok := err.(*PortInUseError); ok {
	return fmt.Sprintf("stop %q or choose a port other than %d", perr.Owner, perr.Port)
}
```
