## Task 1: total minutes first

`float64(days * 24 * 60)` is the total minutes. Multiply by
`(100 - slaPercent) / 100`.

## Task 2: one line

```go
return AllowedDowntimeMinutes(slaPercent, days) - usedMinutes
```

## Task 3: two branches

Store `remaining := RemainingBudget(...)`. If `remaining < 0`, format
`-remaining` with `"%s: budget exceeded by %.1f minutes"`. Otherwise use
`"%s: %.1f of %.1f minutes left"` with the allowed downtime as the second
number.
