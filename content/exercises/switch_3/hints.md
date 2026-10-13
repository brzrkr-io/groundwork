## Task 1: Name the service on a port

Use `switch port` with one case per named port, using commas for rows with
two ports. The "any other port below `1024`" row isn't a single value, so
think about where that check could go.

### Closer

A `default` can hold any code, including an `if`. Put the "below 1024" check
there:

```go
default:
	if port < 1024 {
		return "system"
	}
	return "unknown"
```

### Answer

```go
switch port {
case 22:
	return "ssh"
case 80, 8080:
	return "http"
case 443, 8443:
	return "https"
case 5432:
	return "postgres"
case 6379:
	return "redis"
default:
	if port < 1024 {
		return "system"
	}
	return "unknown"
}
```

## Task 2: Decide what to do with a container

Start with `switch state {` and add one case per state from the table. Most
cases just return. The `"exited"` case needs to look at `exitCode` and
`restarts` too.

### Closer

Inside `case "exited":`, put a tagless `switch {` that checks the rows in
order: `exitCode == 0` first, then `restarts >= 5`, then `default`.

### Answer

```go
switch state {
case "running":
	return "none"
case "exited":
	switch {
	case exitCode == 0:
		return "none"
	case restarts >= 5:
		return "give up"
	default:
		return "restart"
	}
case "oomkilled":
	return "add memory"
default:
	return "inspect"
}
```
