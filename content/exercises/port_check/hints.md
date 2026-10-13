## Task 1: check in order

Check `err == nil` first, then connection refused, then timeout, and fall
through to `Failed`. For the timeout check, use
`errors.AsType[net.Error](err)` and then call `Timeout()` on the result.

## Task 2: a method value

Build the dialer, then store its method without calling it:

```go
d := &net.Dialer{Timeout: timeout}
return &Checker{Dial: d.DialContext}
```

In `Check`, `conn` is only usable when `err` is `nil`, so close it inside an
`if err == nil` block.

## Task 3: build the string line by line

Write into a `strings.Builder` with `fmt.Fprintf`: the header, one row per
result, and the summary. Count the open ones as you go:

```go
var b strings.Builder
fmt.Fprintf(&b, "%-21s %s\n", "ADDRESS", "STATUS")
open := 0
for _, r := range results {
	fmt.Fprintf(&b, "%-21s %s\n", r.Addr, r.Status)
	// count r.Status == Open here
}
fmt.Fprintf(&b, "%d of %d ports open\n", open, len(results))
return b.String()
```

## Almost the answer for Check

```go
conn, err := c.Dial(context.Background(), "tcp", addr)
if err == nil {
	conn.Close()
}
return Result{Addr: addr, Status: Classify(err)}
```
