# Functions: SLO error budget report

An *SLA* (service level agreement) promises uptime, like 99.9%. The
leftover 0.1% is your *error budget*: the downtime you're allowed. SRE teams
track it to decide when to slow down releases. You'll build a small budget
report out of three functions, each one using the one before it.

## Learn

### Recap

Split a job into small functions with clear names, and let each function call
the helpers below it. Convert `int` to `float64` with `float64(n)` before
mixing it with floats, and format floats with `%.1f` for one decimal place.

Floats are not always exact: `100 - 99.9` is `0.09999999999999432`, not
`0.1`. That's normal. The tests compare your answers after formatting them
with `%.1f`, so tiny errors like this don't matter.

## Your task

### 1. Allowed downtime

Write `AllowedDowntimeMinutes(slaPercent float64, days int) float64`. A day
has `24 * 60` minutes. The allowed downtime is the total minutes in `days`
times the share that is *not* promised: `(100 - slaPercent) / 100`.

```go
AllowedDowntimeMinutes(99.9, 30) // => 43.2  (43200 minutes × 0.001)
AllowedDowntimeMinutes(99.0, 7)  // => 100.8
```

### 2. Remaining budget

Write `RemainingBudget(slaPercent float64, days int, usedMinutes float64) float64`:
the allowed downtime minus `usedMinutes`. Call `AllowedDowntimeMinutes`. The
result is negative when the budget is blown.

```go
RemainingBudget(99.9, 30, 30) // => 13.2
RemainingBudget(99.9, 30, 50) // => -6.8
```

### 3. Budget report

Write `BudgetReport(service string, slaPercent float64, days int, usedMinutes float64) string`.
Call `RemainingBudget` and `AllowedDowntimeMinutes`. All numbers use one
decimal place.

- Remaining is 0 or more: `<service>: <remaining> of <allowed> minutes left`
- Remaining is below 0: `<service>: budget exceeded by <how much> minutes`,
  where "how much" is a positive number.

```go
BudgetReport("api", 99.9, 30, 30)   // => "api: 13.2 of 43.2 minutes left"
BudgetReport("api", 99.9, 30, 50)   // => "api: budget exceeded by 6.8 minutes"
```

## Learn more

- [Google SRE book: Embracing risk](https://sre.google/sre-book/embracing-risk/)
- [A Tour of Go: Functions](https://go.dev/tour/basics/4)
- [Go by Example: String formatting](https://gobyexample.com/string-formatting)
