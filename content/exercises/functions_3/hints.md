## Task 1: Allowed downtime

Work it out in two steps: how many minutes are in `days` days, and what share
of that time is *not* promised by the SLA. Multiply the two.

### Closer

The total minutes are `days * 24 * 60`. That is an `int`, so wrap it in
`float64(...)` before mixing it with `slaPercent`. Then multiply by
`(100 - slaPercent) / 100`.

### Answer

```go
return float64(days*24*60) * (100 - slaPercent) / 100
```

## Task 2: Remaining budget

You already wrote a function that gives the allowed downtime. Call it, then
take away the minutes that were used.

### Closer

Pass the same `slaPercent` and `days` along:
`AllowedDowntimeMinutes(slaPercent, days)`. Subtract `usedMinutes` from what it
returns.

### Answer

```go
return AllowedDowntimeMinutes(slaPercent, days) - usedMinutes
```

## Task 3: Budget report

First work out the remaining budget with `RemainingBudget`. Then there are two
cases: it is below 0, or it is not. Each case gets its own message.

### Closer

Store it: `remaining := RemainingBudget(slaPercent, days, usedMinutes)`. Use
`if remaining < 0 { ... }`. The fixed words go inside the format string in
quotes; the values go after it, matching each verb: `%s` for `service`, `%.1f`
for numbers. "How much" must be positive, so pass `-remaining`.

### Answer

```go
remaining := RemainingBudget(slaPercent, days, usedMinutes)
if remaining < 0 {
	return fmt.Sprintf("%s: budget exceeded by %.1f minutes", service, -remaining)
}
allowed := AllowedDowntimeMinutes(slaPercent, days)
return fmt.Sprintf("%s: %.1f of %.1f minutes left", service, remaining, allowed)
```
