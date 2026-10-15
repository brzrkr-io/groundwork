# Switch: container supervisor

A container supervisor looks at each container and decides what to do. That
decision is a perfect job for `switch`.

**New here:** code inside a case (an `if`, or a second switch); practice.

## Learn

### Recap

| Kind | Looks like | Use it when |
|---|---|---|
| Expression switch | `switch x {` then `case 1, 2:` | comparing one value with fixed values |
| Tagless switch | `switch {` then `case x > 5:` | checking conditions |

Only the first matching case runs. `default` catches the rest.

### Code inside a case

The body of a case (or of `default`) is normal code. It can hold an `if`, or
even another switch:

```go
switch day {
case "saturday", "sunday":
	return "weekend"
default:
	if hour < 9 {
		return "too early"
	}
	return "workday"
}
```

> **Why is it written like this?** Go compares an expression switch's value
> with each case value using `==`, so a range like "below 1024" can't be a
> case value. Put that check inside a case body, or use a tagless switch.
> Both are normal Go; the choice is yours.

### Order matters

Cases are checked top to bottom, and the first match wins. Put the most
specific check first.

```go
switch {
case score == 100:
	return "perfect"
case score >= 50: // 100 would also match here, but the case above ran first
	return "pass"
}
```

## Your task

You're writing parts of a small container supervisor. The starter functions
only return a placeholder; write the switch in each one.

### 1. Name the service on a port

**Goal:** make `PortService(port)` return the service name for a port.

| `port` | Result |
|---|---|
| `22` | `"ssh"` |
| `80` or `8080` | `"http"` |
| `443` or `8443` | `"https"` |
| `5432` | `"postgres"` |
| `6379` | `"redis"` |
| any other port below `1024` | `"system"` |
| anything else | `"unknown"` |

"Below `1024`" means `1023` is `"system"` and `1024` is `"unknown"`.

```go
PortService(8443) // => "https"
PortService(25)   // => "system"
PortService(9000) // => "unknown"
```

**Think it through:**
- Which rows are single values you can list in cases?
- The "below 1024" row isn't a single value. Where could that check go?

### 2. Decide what to do with a container

**Goal:** make `ContainerAction(state, exitCode, restarts)` return the action for a container.

| `state` | Condition | Result |
|---|---|---|
| `"running"` | | `"none"` |
| `"exited"` | `exitCode` is `0` | `"none"` |
| `"exited"` | `restarts` is `5` or more | `"give up"` |
| `"exited"` | otherwise | `"restart"` |
| `"oomkilled"` | | `"add memory"` |
| anything else (including `""`) | | `"inspect"` |

For `"exited"`, check the rows in the order shown: a clean exit (`0`) is
`"none"` even after many restarts. Only `"exited"` looks at `exitCode` and
`restarts`; the other states ignore them.

```go
ContainerAction("running", 0, 0) // => "none"
ContainerAction("exited", 1, 2)  // => "restart"
ContainerAction("exited", 1, 5)  // => "give up"
ContainerAction("exited", 0, 9)  // => "none"
ContainerAction("paused", 0, 0)  // => "inspect"
```

**Think it through:**
- Which parameter does the outer switch compare?
- Inside the `"exited"` case, which check must come first?

## Learn more

- [A Tour of Go: Switch evaluation order](https://go.dev/tour/flowcontrol/10)
- [Go by Example: Switch](https://gobyexample.com/switch)
- [Effective Go: Switch](https://go.dev/doc/effective_go#switch)
