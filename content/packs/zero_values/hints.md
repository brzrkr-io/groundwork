## Task 1: Declare variables with their zero values

Instead of giving each variable a value, tell Go only its name and its type.
Go fills in the zero value for you.

### Closer

The pattern is `var name type`, for example `var port int`. Replace each of
the four `:=` lines this way.

### Answer

```go
var port int
var host string
var debug bool
var ratio float64
```

## Task 2: Default the port

"Not set" means the port is `0`. Check for that, return the default, and
otherwise return the port you were given.

### Closer

Compare with `==`: `if port == 0 { ... }`. Inside, return `8080`. Keep
`return port` after the `if`.

### Answer

```go
if port == 0 {
	return 8080
}
return port
```

## Task 3: Build a connection string

Fix each parameter in turn by overwriting it, then let the `fmt.Sprintf`
line that's already there format the result.

### Closer

An empty string is `""`, so check `if host == "" {` and set
`host = "localhost"`. For the port, reuse task 2:
`port = PortOrDefault(port)`. Do the same as `host` for `timeoutSeconds`.

### Answer

```go
if host == "" {
	host = "localhost"
}
port = PortOrDefault(port)
if timeoutSeconds == 0 {
	timeoutSeconds = 30
}
return fmt.Sprintf("%s:%d timeout=%ds", host, port, timeoutSeconds)
```
