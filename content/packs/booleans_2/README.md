# Booleans: fix the alert rules

The alerting rules for your team's pager were written in a hurry. Each rule
has a small bug in its boolean expression, so people get woken up for
nothing (or not at all).

**New here:** nothing; practice reading and fixing boolean expressions.

## Learn

### Recap

As in *Booleans*, a comparison such as `diskPercent <= 90` gives a `bool`.
`&&` needs both sides true, `||` needs at least one side true, and `!` flips a
value. You can return a boolean expression directly.

### Reading a rule out loud

When a boolean expression is wrong, read it out loud and compare it with the
rule in words:

```go
return region == "eu-west-1" || !paused
// "region is eu-west-1, OR not paused"
```

Use this table to turn words into operators:

| The rule says | The code needs |
|---|---|
| "and", "both" | `&&` |
| "or", "either" | `\|\|` |
| "not", "no" | `!` in front of the boolean |
| "below", "less than" | `<` |
| "at most", "no more than" | `<=` |

> **Why is it written like this?** The `!` goes directly in front of the
> value it flips, with no space; that's convention (`gofmt` style). Go only
> requires that it comes before the value.

### Fixing, not rewriting

Each bug here is a wrong operator, a missing `!`, or both. Change only those
parts of the `return` line and leave the rest as it is.

## Your task

Fix each function so it matches its rule.

### 1. Page only for unacknowledged critical alerts

**Goal:** make `ShouldPage` return `true` only when `severity` is exactly
`"critical"` **and** the alert has **not** been acknowledged.

```go
ShouldPage("critical", false) // => true
ShouldPage("critical", true)  // => false (someone is already on it)
ShouldPage("warning", false)  // => false
ShouldPage("warning", true)   // => false
```

**Think it through:**
- Read the current line out loud. Does it say "and" or "or"?
- `acknowledged` is `true` when someone *has* seen it. Does the rule want that,
  or the opposite?

### 2. Fix the health check

**Goal:** make `IsHealthy` treat exactly 500 ms latency as healthy.

A service is healthy when `errorRate` is below `0.01` (1%) **and**
`latencyMs` is at most `500`. Exactly `0.01` is **not** healthy.

```go
IsHealthy(0.005, 120) // => true
IsHealthy(0.001, 500) // => true
IsHealthy(0.01, 120)  // => false
IsHealthy(0.2, 120)   // => false
IsHealthy(0.005, 501) // => false
```

**Think it through:**
- The error-rate part is already right. Look at the latency part only.
- Does the current operator allow 500 itself? Check the table.

### 3. Fix the maintenance window

**Goal:** make `InMaintenanceWindow` return `true` for hours 22, 23, 0 and 1.

The nightly window runs from 22:00 to 01:59, so it *wraps* past midnight.
`hour` runs from 0 to 23. An hour is in the window when it is 22 or more,
**or** below 2.

```go
InMaintenanceWindow(22) // => true
InMaintenanceWindow(23) // => true
InMaintenanceWindow(0)  // => true
InMaintenanceWindow(1)  // => true
InMaintenanceWindow(2)  // => false
InMaintenanceWindow(12) // => false
InMaintenanceWindow(21) // => false
```

**Think it through:**
- Can any hour be both 22-or-more and below 2 at the same time?
- The two comparisons are fine. What joins them?

## Learn more

- [Go spec: Logical operators](https://go.dev/ref/spec#Logical_operators)
- [Go spec: Comparison operators](https://go.dev/ref/spec#Comparison_operators)
