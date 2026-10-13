## Task 1: two verbs

The format string needs a `%d` for the port and a `%q` for the owner, and
`fmt.Sprintf` needs both `e.Port` and `e.Owner` after it, in that order.

## Task 2: where does success go?

On the free-port path, `err` is still a nil `*PortInUseError`, and returning
it fills the type part of the `error`. Remove the `var err` line. Return the
`&PortInUseError{...}` straight from the `if`, and end with `return nil`.

## Task 3: comma-ok

Use the two-result type assertion. It never panics:

```go
if perr, ok := err.(*PortInUseError); ok {
	return fmt.Sprintf("stop %q or choose a port other than %d", perr.Owner, perr.Port)
}
```

Put it before the `"unexpected error: "` line.
