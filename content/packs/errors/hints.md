## Task 1: Reject an empty service name

Compare `name` with the empty string. If it matches, return a new error; otherwise there is nothing wrong, so return `nil`.

### Closer

Add `"errors"` to your imports. `errors.New("...")` makes an error with a fixed message, and the check is `if name == "" { ... }`.

### Answer

```go
if name == "" {
	return errors.New("service name is empty")
}
return nil
```

The function ends with `return nil` because reaching that line means the name was fine.

## Task 2: Parse a port number

Convert the text with `strconv.Atoi` and check its error first. Then check the range. Each problem gets its own `if` that returns `0` and an error right away.

### Closer

Use `fmt.Errorf` with `%q` for the text and `%d` for the number:

```go
n, err := strconv.Atoi(s)
if err != nil {
	return 0, fmt.Errorf("port %q is not a number", s)
}
// now check n < 1 || n > 65535
```

### Answer

```go
n, err := strconv.Atoi(s)
if err != nil {
	return 0, fmt.Errorf("port %q is not a number", s)
}
if n < 1 || n > 65535 {
	return 0, fmt.Errorf("port %d is out of range 1-65535", n)
}
return n, nil
```

Each failure returns early with the zero value `0`, so only a good port reaches `return n, nil`.

## Task 3: Parse an endpoint

Split `s` on the first `:` with `strings.Cut`, then check the problems in the README's order: no colon, empty host, bad port. Reuse your `ParsePort` for the last one.

### Closer

```go
host, portText, found := strings.Cut(s, ":")
if !found {
	return "", 0, fmt.Errorf("endpoint %q has no port", s)
}
```

When `ParsePort(portText)` fails, add context with `fmt.Errorf("endpoint %q: %v", s, err)`.

### Answer

```go
host, portText, found := strings.Cut(s, ":")
if !found {
	return "", 0, fmt.Errorf("endpoint %q has no port", s)
}
if host == "" {
	return "", 0, fmt.Errorf("endpoint %q has no host", s)
}
port, err = ParsePort(portText)
if err != nil {
	return "", 0, fmt.Errorf("endpoint %q: %v", s, err)
}
return host, port, nil
```

`%v` puts `ParsePort`'s message after your context, so the reader sees both where and what went wrong.
