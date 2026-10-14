## Task 1: Fix the probe port

Read the error: `ProbePort` is an untyped string constant, but `HealthPort`
must return an `int`. What makes `"8081"` a string instead of a number?

### Closer

The quotes make it a string. Fix the constant, not the function: remove the
quotes around `8081`.

### Answer

```go
ProbePort = 8081
```

## Task 2: Fix the health path

Only the constant's value is wrong; `HealthURL` itself is already right.
Compare `/health` with what the load balancer probes.

### Closer

Change the text inside the quotes of `HealthPath` to `/healthz`. It is one
extra character.

### Answer

```go
HealthPath = "/healthz"
```

## Task 3: Compute the time to mark a target down

`DownAfterSeconds` should be the interval times the number of failed checks.
Use the constant names, not the number 30, so it stays right if either changes.

### Closer

`*` multiplies. A constant's value can use other constants:
`DownAfterSeconds = CheckIntervalSeconds * ...`.

### Answer

```go
DownAfterSeconds = CheckIntervalSeconds * FailureThreshold
```
