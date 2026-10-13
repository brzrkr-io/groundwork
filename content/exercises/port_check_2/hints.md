## Task 1: let net do the brackets

`net.JoinHostPort` takes the port as a string, so convert it first:
`net.JoinHostPort(host, strconv.Itoa(port))`.

## Task 2: one check at a time

Split first, then check each rule and return early when one fails:

```go
host, portText, err := net.SplitHostPort(s)
if err != nil {
	return "", 0, err
}
if host == "" {
	return "", 0, fmt.Errorf("target %q has no host", s)
}
```

Then convert `portText` with `strconv.Atoi` and check `port < 1 || port > 65535`.

## Task 3: return and sleep

Put `return nil` inside the `if err == nil` block, after `conn.Close()`. For
the sleep, compare the loop counter with `attempts`:

```go
if attempt < attempts {
	sleep(time.Second)
}
```
