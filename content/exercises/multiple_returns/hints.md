## Two values, two expressions

For task 1, hours are `totalMinutes / 60` and the leftover minutes are
`totalMinutes % 60`. Return both: `return hours, minutes`.

## Task 2: use all three results of strings.Cut

```go
host, port, found := strings.Cut(addr, ":")
if !found {
	return "", "", false
}
```

## Task 3: reuse SplitHostPort

Call your `SplitHostPort` first. If it says not found, return
`"", 0, errors.New("missing port")`. Then convert the port with
`strconv.Atoi` and check `err != nil`.

## Almost the answer

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
