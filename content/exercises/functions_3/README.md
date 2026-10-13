# Functions: SLO error budget report

An *SLA* (service level agreement) promises uptime, like 99.9%. The
leftover 0.1% is your *error budget*: the downtime you're allowed. SRE teams
track it to decide when to slow down releases. You'll build a small budget
report out of three functions, each one using the one before it.

**New here:** nothing; practice with functions, `float64` and `%.1f`.

## Learn

### Recap: functions that build on each other

Split a job into small functions. Each one calls the helper below it and
passes its own parameters along:

```go
func Area(w, h float64) float64 { return w * h }

func Paint(w, h, coats float64) float64 {
	return Area(w, h) * coats
}
```

`Paint` doesn't redo the maths; it calls `Area` with its own `w` and `h`.

### Mixing `int` and `float64`

Go won't multiply an `int` by a `float64`. Convert first with `float64(n)`:

```go
hours := 3                        // int
cost := float64(hours) * 12.5     // => 37.5
```

> **Why is it written like this?** Go requires the conversion; it never
> converts number types for you. Where you put it is your choice, as long as
> both sides of each operator end up the same type.

### Negative numbers and `-x`

Putting `-` in front of a value flips its sign: if `x` is `-6.8`, then `-x`
is `6.8`. Use `<` with `0` to check whether a number is negative.

### Floats are not always exact

`100 - 99.9` is `0.09999999999999432`, not `0.1`. That's normal. The tests
accept any answer within 0.05 of the expected one, and the report compares
text formatted with `%.1f`, so tiny errors don't matter.

## Your task

### 1. Allowed downtime

**Goal:** make `AllowedDowntimeMinutes(slaPercent float64, days int) float64` return how many minutes of downtime the SLA allows over `days`.

A day has `24 * 60` minutes. Multiply the total minutes in `days` by the
share that is *not* promised: `(100 - slaPercent) / 100`.

```go
AllowedDowntimeMinutes(99.9, 30) // => 43.2  (43200 minutes × 0.001)
AllowedDowntimeMinutes(99.0, 7)  // => 100.8
AllowedDowntimeMinutes(99.5, 1)  // => 7.2
```

**Think it through:**
- `days` is an `int` but `slaPercent` is a `float64`. Where do you convert?
- Which parameter gives the total minutes, and which gives the share?

### 2. Remaining budget

**Goal:** make `RemainingBudget(slaPercent float64, days int, usedMinutes float64) float64` return the allowed downtime minus `usedMinutes`.

Get the allowed downtime by calling `AllowedDowntimeMinutes`. The result is
negative when the budget is blown.

```go
RemainingBudget(99.9, 30, 30)  // => 13.2
RemainingBudget(99.9, 30, 50)  // => -6.8
RemainingBudget(99.0, 7, 0.8)  // => 100.0
```

**Think it through:**
- Which two of your parameters does `AllowedDowntimeMinutes` need?

### 3. Budget report

**Goal:** make `BudgetReport(service string, slaPercent float64, days int, usedMinutes float64) string` return one line describing the budget.

Use `RemainingBudget` and `AllowedDowntimeMinutes`. Every number has one
decimal place. `service` comes from the parameter; the rest of the words are
fixed text.

| Remaining is… | Line |
|---|---|
| 0 or more | `<service>: <remaining> of <allowed> minutes left` |
| below 0 | `<service>: budget exceeded by <how much> minutes` |

"How much" is the overspend as a positive number (remaining `-6.8` prints `6.8`).

```go
BudgetReport("api", 99.9, 30, 30) // => "api: 13.2 of 43.2 minutes left"
BudgetReport("web", 99.0, 7, 0)   // => "web: 100.8 of 100.8 minutes left"
BudgetReport("api", 99.9, 30, 50) // => "api: budget exceeded by 6.8 minutes"
BudgetReport("db", 99.5, 1, 10)   // => "db: budget exceeded by 2.8 minutes"
```

**Think it through:**
- Store the remaining budget in a variable first, so you can check it and print it.
- Which verb prints a string, and which prints one decimal place?
- How do you turn a negative number into its positive version?

## Learn more

- [Google SRE book: Embracing risk](https://sre.google/sre-book/embracing-risk/)
- [A Tour of Go: Functions](https://go.dev/tour/basics/4)
- [Go by Example: String formatting](https://gobyexample.com/string-formatting)
