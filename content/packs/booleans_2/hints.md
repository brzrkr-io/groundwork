## Task 1: Page only for unacknowledged critical alerts

Read the code out loud and compare it with the rule in the README. The rule
says "and" and "not acknowledged". Does the code say the same?

### Closer

Two fixes: swap `||` for `&&`, and put `!` in front of `acknowledged` so it
means "not acknowledged".

### Answer

```go
return severity == "critical" && !acknowledged
```

## Task 2: Fix the health check

The error rate part is fine. Look at the latency part: should exactly 500 be
healthy? Does `<` allow 500?

### Closer

"At most 500" means "less than or equal to 500". Change one operator in
`latencyMs < 500`.

### Answer

```go
return errorRate < 0.01 && latencyMs <= 500
```

## Task 3: Fix the maintenance window

Can any hour be both 22 or more **and** below 2? The window is late at night
**or** early in the morning.

### Closer

Keep both comparisons, but swap `&&` for `||`.

### Answer

```go
return hour >= 22 || hour < 2
```

The window wraps past midnight, so an hour is in it when it is late or early,
never both.
