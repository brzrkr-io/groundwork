## Task 1: Classify the dial error

Check the cases in order: no error, connection refused, timeout, and anything
else last. The Learn section shows both error checks.

### Closer

`errors.Is(err, syscall.ECONNREFUSED)` spots a closed port. For the timeout,
`errors.AsType[net.Error](err)` returns the error and an `ok`; then call
`Timeout()` on it.

### Answer

```go
if err == nil {
	return Open
}
if errors.Is(err, syscall.ECONNREFUSED) {
	return Closed
}
if netErr, ok := errors.AsType[net.Error](err); ok && netErr.Timeout() {
	return Timeout
}
return Failed
```

## Task 2: Dial and check

`NewChecker` stores a method value: the dialer's `DialContext` method without
calling it. `Check` calls `c.Dial` and lets `Classify` decide the status.

### Closer

`d := &net.Dialer{Timeout: timeout}` and `Dial: d.DialContext` (no `()`). In
`Check`, `conn` is only usable when `err` is `nil`, so close it inside an
`if err == nil` block.

### Answer

```go
func NewChecker(timeout time.Duration) *Checker {
	d := &net.Dialer{Timeout: timeout}
	return &Checker{Dial: d.DialContext}
}

func (c *Checker) Check(addr string) Result {
	conn, err := c.Dial(context.Background(), "tcp", addr)
	if err == nil {
		conn.Close()
	}
	return Result{Addr: addr, Status: Classify(err)}
}
```

## Task 3: Print a report

Write the header, one row per result and the summary into a
`strings.Builder`. Count the open ports as you go.

### Closer

Every row uses the same format: `fmt.Fprintf(&b, "%-21s %s\n", r.Addr, r.Status)`.
The summary is `"%d of %d ports open\n"`.

### Answer

```go
var b strings.Builder
fmt.Fprintf(&b, "%-21s %s\n", "ADDRESS", "STATUS")
open := 0
for _, r := range results {
	fmt.Fprintf(&b, "%-21s %s\n", r.Addr, r.Status)
	if r.Status == Open {
		open++
	}
}
fmt.Fprintf(&b, "%d of %d ports open\n", open, len(results))
return b.String()
```
