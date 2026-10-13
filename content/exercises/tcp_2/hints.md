## Task 1: Fix the stats server

The client reads with `io.ReadAll`, which only stops when the connection
closes. Nothing in `HandleStats` ever closes it.

### Closer

A `defer` at the top of the function runs however the function returns.

### Answer

```go
defer conn.Close()
```

Put it as the first line of `HandleStats`.

## Task 2: Recognise a timeout

Use `errors.As` to look for a `net.Error` inside `err`, then ask it whether it
was a timeout. `errors.As` already returns `false` for `nil`.

### Closer

Declare a variable of the interface type and pass its address:
`var netErr net.Error`, then `errors.As(err, &netErr)`.

### Answer

```go
var netErr net.Error
return errors.As(err, &netErr) && netErr.Timeout()
```

## Task 3: Finish the banner grabber

The deadline must be set *before* the read, or a silent server makes the read
wait forever. The banner also still carries its line ending.

### Closer

`conn.SetReadDeadline(time.Now().Add(timeout))` goes before `ReadString`.
`strings.TrimRight(line, "\r\n")` removes both `\r` and `\n`.

### Answer

```go
if err := conn.SetReadDeadline(time.Now().Add(timeout)); err != nil {
	return "", err
}
line, err := bufio.NewReader(conn).ReadString('\n')
if err != nil {
	return "", err
}
return strings.TrimRight(line, "\r\n"), nil
```
