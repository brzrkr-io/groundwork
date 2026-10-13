## Task 1: Split uptime into hours and minutes

An hour has 60 minutes. Whole hours come from dividing, and the leftover
minutes come from the remainder. The function now returns two values.

### Closer

`totalMinutes / 60` gives the whole hours, and `totalMinutes % 60` gives the
minutes left over. Return both, separated by a comma: `return a, b`.

### Answer

```go
return totalMinutes / 60, totalMinutes % 60
```

## Task 2: Split a host and a port

`strings.Cut` hands back three things: the part before the `:`, the part after
it, and whether it found a `:` at all. Use all three.

### Closer

Receive the three results: `host, port, found := strings.Cut(addr, ":")`. If
`!found`, return `"", "", false`. Otherwise return `host`, `port` and `true`.

### Answer

```go
host, port, found := strings.Cut(addr, ":")
if !found {
	return "", "", false
}
return host, port, true
```

## Task 3: Parse an address

Reuse your `SplitHostPort` to get the host and the port text. Then turn the
port text into a number, and check each thing that can go wrong, one `if` at a
time.

### Closer

If `SplitHostPort` says not ok, return `"", 0, errors.New("missing port")`.
Convert with `port, err := strconv.Atoi(portText)` and, if `err != nil`, return
`"", 0, err`. Last, check `port < 1 || port > 65535`. Add `errors`, `fmt` and
`strconv` to the imports.

### Answer

```go
host, portText, ok := SplitHostPort(addr)
if !ok {
	return "", 0, errors.New("missing port")
}
port, err := strconv.Atoi(portText)
if err != nil {
	return "", 0, err
}
if port < 1 || port > 65535 {
	return "", 0, fmt.Errorf("port %d out of range", port)
}
return host, port, nil
```
