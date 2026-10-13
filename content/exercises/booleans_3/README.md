# Booleans: autoscaler decisions

An *autoscaler* adds replicas of a service when it is busy and removes them
when it is quiet. You will write the three yes-or-no decisions at its core.

## Learn

### Recap

Comparisons (`<`, `>`, `<=`, `>=`, `==`, `!=`) give a `bool`, and `&&`, `||`
and `!` combine booleans, as in *Booleans*. A function can return a boolean
expression directly.

### Calling your own checks

A call to a function that returns `bool` is itself a boolean expression, so
you can combine calls with `&&`, `||` and `!`:

```go
return !IsPrivilegedPort(port) && port != 8080
```

This keeps each rule in one place: if the rule in `IsPrivilegedPort` changes,
every caller picks up the change.

### Cooldown

After a scaling action, an autoscaler waits for a *cooldown* period before it
acts again, so it doesn't flap up and down. While `cooldown` is `true`, it
makes no changes at all.

## Your task

All three functions currently return `false`. Write the real decisions.

### 1. Decide when to scale up

`ShouldScaleUp` returns `true` when **all** of these hold:

- `cpuPercent` is above 75 (75 itself is not enough),
- `replicas` is below `maxReplicas`, so there is room to grow,
- no cooldown is active.

```go
ShouldScaleUp(90, 3, 10, false)  // => true
ShouldScaleUp(75, 3, 10, false)  // => false (not above 75)
ShouldScaleUp(90, 10, 10, false) // => false (already at the maximum)
ShouldScaleUp(90, 3, 10, true)   // => false (cooldown)
```

### 2. Decide when to scale down

`ShouldScaleDown` returns `true` when **all** of these hold:

- `cpuPercent` is below 25 (25 itself is not low enough),
- `replicas` is above `minReplicas`, so there is something to remove,
- no cooldown is active.

```go
ShouldScaleDown(10, 5, 2, false) // => true
ShouldScaleDown(25, 5, 2, false) // => false (not below 25)
ShouldScaleDown(10, 2, 2, false) // => false (already at the minimum)
ShouldScaleDown(10, 5, 2, true)  // => false (cooldown)
```

### 3. Decide when the service is stable

`IsStable` returns `true` when the autoscaler would do nothing: it should
**not** scale up **and** it should **not** scale down. Build it by calling
`ShouldScaleUp` and `ShouldScaleDown` with the same inputs, rather than
repeating their rules.

```go
IsStable(50, 3, 2, 10, false) // => true (CPU is fine)
IsStable(90, 3, 2, 10, false) // => false (would scale up)
IsStable(10, 5, 2, 10, false) // => false (would scale down)
IsStable(90, 3, 2, 10, true)  // => true (cooldown: no action)
```

## Learn more

- [Go spec: Logical operators](https://go.dev/ref/spec#Logical_operators)
- [Kubernetes: Horizontal Pod Autoscaling](https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale/)
