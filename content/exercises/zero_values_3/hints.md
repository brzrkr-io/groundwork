## Task 1: Pick the effective timeout

Check the layers one at a time, in order: request, then service, then
global. As soon as one isn't `0`, that's your answer, so return it straight
away.

### Closer

Write one `if` per layer, using `!= 0`. If none of them returned, the
function reaches the end, where you return `30`:

```go
if requestSeconds != 0 {
	return requestSeconds
}
// ...the same for serviceSeconds and globalSeconds...
return 30
```

### Answer

```go
if requestSeconds != 0 {
	return requestSeconds
}
if serviceSeconds != 0 {
	return serviceSeconds
}
if globalSeconds != 0 {
	return globalSeconds
}
return 30
```

## Task 2: Describe a health probe

Do it in two steps. First, replace each parameter that holds its zero value
with its default. Then build the string once, at the end.

### Closer

For each default, use `if x == zero { x = default }`, for example
`if port == 0 { port = 8080 }`. In the `fmt.Sprintf` format, the fixed text
(`GET `, ` on :`, ` every `, `s`) goes inside the quotes. The parameters
don't: they go after the format, matched to `%s` (string) and `%d` (int).

### Answer

```go
if path == "" {
	path = "/healthz"
}
if port == 0 {
	port = 8080
}
if periodSeconds == 0 {
	periodSeconds = 10
}
return fmt.Sprintf("GET %s on :%d every %ds", path, port, periodSeconds)
```
