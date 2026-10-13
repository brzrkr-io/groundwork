## Task 1: Fix `Address` for IPv6

Don't add the brackets by hand. The `net` package has a function that joins a
host and a port and knows about IPv6.

### Closer

`net.JoinHostPort(host, port)` takes the port as a string, so convert the
`int` first with `strconv.Itoa(port)`.

### Answer

```go
return net.JoinHostPort(host, strconv.Itoa(port))
```

## Task 2: Finish `ParseTarget`

Split the target first, then check one rule at a time and return an error as
soon as a rule fails. `net.SplitHostPort` does the splitting but does not
check the host or the port number for you.

### Closer

Start with `host, portText, err := net.SplitHostPort(s)` and return on error.
Then check `host == ""`, convert with `strconv.Atoi(portText)`, and check
`port < 1 || port > 65535`.

### Answer

```go
host, portText, err := net.SplitHostPort(s)
if err != nil {
	return "", 0, err
}
if host == "" {
	return "", 0, fmt.Errorf("target %q has no host", s)
}
port, err := strconv.Atoi(portText)
if err != nil {
	return "", 0, fmt.Errorf("target %q: port is not a number", s)
}
if port < 1 || port > 65535 {
	return "", 0, fmt.Errorf("target %q: port %d is outside 1-65535", s, port)
}
return host, port, nil
```

## Task 3: Finish `WaitForPort`

A successful dial means you are done, so leave the function right there. For
the sleep, compare the loop counter `attempt` with `attempts` to tell whether
this was the last try.

### Closer

Put `return nil` after `conn.Close()`. Wrap the sleep in
`if attempt < attempts { ... }` and call `sleep(time.Second)` inside.

### Answer

```go
if err == nil {
	conn.Close()
	return nil
}
lastErr = err
if attempt < attempts {
	sleep(time.Second)
}
```
