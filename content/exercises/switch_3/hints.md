## Task 1: default can hold an if

List the known ports with commas, then handle "below 1024" inside `default`:

```go
default:
	if port < 1024 {
		return "system"
	}
	return "unknown"
```

## Task 2: switch on state

Start with `switch state {` and add a case for each state. The `"exited"`
case needs more checks.

## Task 2: the exited case

```go
case "exited":
	switch {
	case exitCode == 0:
		return "none"
	case restarts >= 5:
		return "give up"
	default:
		return "restart"
	}
```
