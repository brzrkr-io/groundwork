## Task 1: Pick an instance size

Each row of the README table becomes one `case` in a `switch env`. The
"anything else" row is the `default`.

### Closer

Strings go in quotes after `case`. For the shared row, list both values with
a comma: `case "dev", "test":`. Each case returns its size.

### Answer

```go
switch env {
case "production":
	return "large"
case "staging":
	return "medium"
case "dev", "test":
	return "small"
default:
	return "unknown"
}
```

## Task 2: Classify an HTTP method

If you turn the method into uppercase first, `"get"`, `"Get"` and `"GET"` all
become the same thing. Then you only need to list uppercase names.

### Closer

Add `import "strings"` below the `package` line. Start the switch with
`switch strings.ToUpper(method) {`, then write two cases with comma-separated
uppercase methods, plus a `default`.

### Answer

```go
switch strings.ToUpper(method) {
case "GET", "HEAD", "OPTIONS":
	return "read"
case "POST", "PUT", "PATCH", "DELETE":
	return "write"
default:
	return "unknown"
}
```

## Task 3: Rate a response time

Use a tagless `switch {` where each `case` is a condition like `ms < 100`.
Cases are checked top to bottom, so put the smallest limit first.

### Closer

Because `case ms < 100:` comes first, the next case `ms < 500` only sees
values from 100 upwards. Keep going with `ms < 2000`, and let `default`
handle everything that's left:

```go
switch {
case ms < 100:
	return "fast"
case ms < 500:
	return "ok"
// ...
}
```

### Answer

```go
switch {
case ms < 100:
	return "fast"
case ms < 500:
	return "ok"
case ms < 2000:
	return "slow"
default:
	return "critical"
}
```
