# Switch: container supervisor

A container supervisor looks at each container and decides what to do. That
decision is a perfect job for `switch`.

## Learn

### Recap

An expression switch (`switch x { case 1, 2: ... }`) compares one value with
each case. A tagless switch (`switch { case x > 5: ... }`) checks conditions
top to bottom. Only the first matching case runs, and `default` catches the
rest. Inside a case you can use any code, including `if` or another `switch`.

## Your task

You're writing parts of a small container supervisor.

### 1. Name the service on a port

`PortService(port)` returns:

| `port` | Result |
|---|---|
| `22` | `"ssh"` |
| `80` or `8080` | `"http"` |
| `443` or `8443` | `"https"` |
| `5432` | `"postgres"` |
| `6379` | `"redis"` |
| any other port below `1024` | `"system"` |
| anything else | `"unknown"` |

```go
PortService(8443) // => "https"
PortService(25)   // => "system"
PortService(9000) // => "unknown"
```

### 2. Decide what to do with a container

`ContainerAction(state, exitCode, restarts)` returns the action to take:

| `state` | Condition | Result |
|---|---|---|
| `"running"` | | `"none"` |
| `"exited"` | `exitCode` is `0` | `"none"` |
| `"exited"` | `restarts` is `5` or more | `"give up"` |
| `"exited"` | otherwise | `"restart"` |
| `"oomkilled"` | | `"add memory"` |
| anything else | | `"inspect"` |

For `"exited"`, check the rows in the order shown: a clean exit (`0`) is
`"none"` even after many restarts.

```go
ContainerAction("running", 0, 0) // => "none"
ContainerAction("exited", 1, 2)  // => "restart"
ContainerAction("exited", 1, 5)  // => "give up"
ContainerAction("exited", 0, 9)  // => "none"
ContainerAction("paused", 0, 0)  // => "inspect"
```

## Learn more

- [A Tour of Go: Switch evaluation order](https://go.dev/tour/flowcontrol/10)
- [Go by Example: Switch](https://gobyexample.com/switch)
