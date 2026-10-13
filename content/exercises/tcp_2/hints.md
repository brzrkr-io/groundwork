## Task 1: one line

The client only stops reading when the connection closes. Add
`defer conn.Close()` as the first line of `HandleStats`.

## Task 2: errors.As with an interface

Declare a variable of the interface type, pass its address to `errors.As`,
and only call `Timeout()` if `errors.As` found one:

```go
var netErr net.Error
return errors.As(err, &netErr) && netErr.Timeout()
```

## Task 3: deadline before the read

The deadline must be set *before* `ReadString`, or the read can wait forever:

```go
if err := conn.SetReadDeadline(time.Now().Add(timeout)); err != nil {
	return "", err
}
```

Then return `strings.TrimRight(line, "\r\n")` instead of `line`.
