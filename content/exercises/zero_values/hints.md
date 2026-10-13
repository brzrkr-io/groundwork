## Task 1: var without a value

Write the type instead of the value: `var name type`. For example
`var port int`. Do this for all four variables.

## Task 2: compare with zero

Use `if port == 0 { return 8080 }`, then return `port` after the `if`.

## Task 3: overwrite, then format

Fix each parameter in turn, then format once:

```go
if host == "" {
	host = "localhost"
}
port = PortOrDefault(port)
// same idea for timeoutSeconds
return fmt.Sprintf("%s:%d timeout=%ds", host, port, timeoutSeconds)
```
