## Task 1: errors.New

Add `import "errors"` and return `errors.New("service name is empty")` when the
name is empty. At the end of the function, return `nil`.

## Task 2: check, then return early

Call `strconv.Atoi(s)`. If it returns an error, return `0` and your own error
from `fmt.Errorf` with `%q`. Then check `n < 1 || n > 65535` the same way,
using `%d`.

```go
n, err := strconv.Atoi(s)
if err != nil {
	return 0, fmt.Errorf("port %q is not a number", s)
}
```

## Task 3: strings.Cut and context

```go
host, portText, found := strings.Cut(s, ":")
if !found {
	return "", 0, fmt.Errorf("endpoint %q has no port", s)
}
```

Then check `host == ""`, then call `ParsePort(portText)` and, on error, return
`fmt.Errorf("endpoint %q: %v", s, err)`.
