## Task 1: Measure the SLI

Handle "no requests" first, then divide good by total. Integer division
would round `999 / 1000` down to 0, so both sides must be floats.

### Closer

Start with `if total == 0 { return ... }`. Then convert both counts:
`float64(good) / ...`.

### Answer

```go
if total == 0 {
	return 1
}
return float64(good) / float64(total)
```

## Task 2: Count the error budget

The share allowed to fail is `1 - target`. `AllowedFailures` multiplies the
total by it. `BudgetRemaining` counts the failures (`total - good`) and asks
what share of the allowance is still unused.

### Closer

`AllowedFailures` is one line: `float64(total) * (...)`. In
`BudgetRemaining`, return 1 when `total == 0`, then
`allowed := AllowedFailures(target, total)`, `failed := float64(total - good)`
and divide the difference by `allowed`.

### Answer

```go
// AllowedFailures
return float64(total) * (1 - target)

// BudgetRemaining
if total == 0 {
	return 1
}
allowed := AllowedFailures(target, total)
failed := float64(total - good)
return (allowed - failed) / allowed
```

## Task 3: Compute the burn rate

The error rate is the part of the SLI that isn't good: `1 - SLI(...)`.
Divide it by the error rate the target allows.

### Closer

Call `SLI(good, total)` and subtract it from 1, then divide by
`(1 - target)`. Mind the parentheses: division happens before subtraction.

### Answer

```go
return (1 - SLI(good, total)) / (1 - target)
```

`SLI` returns 1 for no requests, so the error rate, and the burn rate, are 0.

## Task 4: Decide whether to page

Work out the burn rate of each window with `BurnRate`, and page only when
both are above the threshold.

### Closer

`BurnRate(target, long.Good, long.Total) > threshold` is the long window's
check. Combine it with the short window's check using `&&`.

### Answer

```go
return BurnRate(target, long.Good, long.Total) > threshold &&
	BurnRate(target, short.Good, short.Total) > threshold
```
