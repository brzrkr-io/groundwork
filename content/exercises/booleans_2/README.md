# Booleans: fix the alert rules

The alerting rules for your team's pager were written in a hurry. Each rule
has one small bug in its boolean expression, and people are getting woken up
for nothing (or not at all).

## Learn

### Recap

As in *Booleans*, a comparison such as `latencyMs <= 500` gives a `bool`.
`&&` needs both sides true, `||` needs at least one side true, and `!` flips a
value. You can return a boolean expression directly.

### Reading a rule out loud

When a boolean expression is wrong, read it out loud and compare it with the
rule in words:

```go
return env == "production" || !paused
// "env is production, OR not paused"
```

If the rule says "and", the code needs `&&`. If it says "not", look for a `!`.
If it says "at most", that is `<=`; "below" is `<`.

## Your task

Fix each function so it matches its rule. Each one needs only a small change:
one operator, or one missing `!`.

### 1. Page only for unacknowledged critical alerts

`ShouldPage` should return `true` when `severity` is exactly `"critical"`
**and** the alert has **not** been acknowledged.

```go
ShouldPage("critical", false) // => true
ShouldPage("critical", true)  // => false (someone is already on it)
ShouldPage("warning", false)  // => false
```

### 2. Fix the health check

A service is healthy when its `errorRate` is below `0.01` (1%) **and** its
`latencyMs` is at most `500`. Exactly 500 ms still counts as healthy;
exactly `0.01` does not.

```go
IsHealthy(0.005, 120) // => true
IsHealthy(0.001, 500) // => true
IsHealthy(0.01, 120)  // => false
IsHealthy(0.005, 501) // => false
```

### 3. Fix the maintenance window

The nightly maintenance window starts at 22:00 and ends at 01:59, so it
*wraps* past midnight. `hour` runs from 0 to 23. The window covers hours 22,
23, 0 and 1: `hour` is 22 or more, **or** `hour` is below 2. No hour can be
both, so the current `&&` version is never true.

```go
InMaintenanceWindow(23) // => true
InMaintenanceWindow(1)  // => true
InMaintenanceWindow(2)  // => false
InMaintenanceWindow(12) // => false
```

## Learn more

- [Go spec: Logical operators](https://go.dev/ref/spec#Logical_operators)
- [Go spec: Comparison operators](https://go.dev/ref/spec#Comparison_operators)
