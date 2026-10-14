# Booleans: autoscaler decisions

An *autoscaler* adds replicas of a service when it is busy and removes them
when it is quiet. You will write the three yes-or-no decisions at its core.

**New here:** calling your own `bool` functions inside a boolean expression.

## Learn

### Recap

Comparisons (`<`, `>`, `<=`, `>=`, `==`, `!=`) give a `bool`, and `&&`, `||`
and `!` combine booleans, as in *Booleans*. A function can return a boolean
expression directly.

### Comparing two parameters

A comparison doesn't need a fixed number on one side. It can compare two
variables:

```go
return used < quota // true when there is space left
```

### Calling your own checks

A call to a function that returns `bool` is itself a boolean, so you can
combine calls with `&&`, `||` and `!`:

```go
return !IsPrivilegedPort(port) && port != 8080
```

This keeps each rule in one place: if the rule in `IsPrivilegedPort` changes,
every caller picks up the change.

> **Why is it written like this?** A call needs the function name **and**
> parentheses with the values to pass in, in the order the function lists its
> parameters. That's what Go requires. Using the bare name `IsPrivilegedPort`
> without `(...)` doesn't give you a `bool`, and the compiler rejects it.

### Shared parameter types

```go
func InRange(value, low, high int) bool
```

`value, low, high int` means all three are `int`. That's a shorthand Go
allows; writing `value int, low int, high int` means the same thing.

### Cooldown

After a scaling action, an autoscaler waits for a *cooldown* period before it
acts again, so it doesn't flap up and down. While `cooldown` is `true`, it
makes no changes at all.

## Your task

All three functions currently return `false`. Write the real decisions.

### 1. Decide when to scale up

**Goal:** make `ShouldScaleUp` return `true` when CPU is high, there is room
to grow, and no cooldown is active.

All of these must hold:

- `cpuPercent` is above 75 (75 itself is not enough),
- `replicas` is below `maxReplicas`,
- `cooldown` is `false`.

```go
ShouldScaleUp(90, 3, 10, false)  // => true
ShouldScaleUp(76, 9, 10, false)  // => true
ShouldScaleUp(75, 3, 10, false)  // => false (not above 75)
ShouldScaleUp(40, 3, 10, false)  // => false
ShouldScaleUp(90, 10, 10, false) // => false (already at the maximum)
ShouldScaleUp(90, 3, 10, true)   // => false (cooldown)
```

**Think it through:**
- Each bullet is one small comparison. What joins "all of these"?
- The second bullet compares two parameters, not a parameter and a number.

### 2. Decide when to scale down

**Goal:** make `ShouldScaleDown` return `true` when CPU is low, there is a
replica to remove, and no cooldown is active.

All of these must hold:

- `cpuPercent` is below 25 (25 itself is not low enough),
- `replicas` is above `minReplicas`,
- `cooldown` is `false`.

```go
ShouldScaleDown(10, 5, 2, false) // => true
ShouldScaleDown(24, 3, 2, false) // => true
ShouldScaleDown(25, 5, 2, false) // => false (not below 25)
ShouldScaleDown(80, 5, 2, false) // => false
ShouldScaleDown(10, 2, 2, false) // => false (already at the minimum)
ShouldScaleDown(10, 5, 2, true)  // => false (cooldown)
```

**Think it through:**
- This is the mirror image of task 1. Which operators flip direction?

### 3. Decide when the service is stable

**Goal:** make `IsStable` return `true` when the autoscaler would **not**
scale up **and** would **not** scale down.

Build it by calling `ShouldScaleUp` and `ShouldScaleDown` with the
parameters `IsStable` received, rather than repeating their rules.

```go
IsStable(50, 3, 2, 10, false)  // => true (CPU is fine)
IsStable(90, 3, 2, 10, false)  // => false (would scale up)
IsStable(10, 5, 2, 10, false)  // => false (would scale down)
IsStable(90, 10, 2, 10, false) // => true (busy, but at the maximum)
IsStable(10, 2, 2, 10, false)  // => true (quiet, but at the minimum)
IsStable(90, 3, 2, 10, true)   // => true (cooldown: no action)
```

**Think it through:**
- Look at each function's parameter list. Which of `IsStable`'s parameters
  does each one need, and in what order? (One needs `maxReplicas`, the other
  `minReplicas`.)
- How do you turn "should scale up" into "should **not** scale up"?

## Learn more

- [Go spec: Logical operators](https://go.dev/ref/spec#Logical_operators)
- [A Tour of Go: Functions continued](https://go.dev/tour/basics/5)
- [Kubernetes: Horizontal Pod Autoscaling](https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale/)
