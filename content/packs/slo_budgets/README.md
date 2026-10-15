# SLOs: error budgets and burn rates

An *SLO* (service level objective) is the reliability your team promises,
like "99.9% of requests succeed". You'll measure it from request counts, work
out how much failure you can still afford, and decide when that failure is
burning fast enough to wake someone up.

**New here:** SLIs from request counts, error budgets in requests, burn rates, multi-window burn-rate alerts.

## Learn

### Recap

In *Functions: SLO error budget report* you turned an uptime percentage into
minutes of allowed downtime. That works for "is it up?". Busy APIs are
measured by **requests** instead, because "up" can still mean 3% of
requests fail.

### SLI: the measurement

An *SLI* (service level indicator) is the share of requests that were good:

```text
SLI = good requests ÷ total requests
```

What counts as "good" is your choice: often "not a 5xx and faster than
300ms". An SLI is a fraction between 0 and 1: `0.998` means 99.8% good.

With **no requests at all** (total 0), nothing failed. This exercise treats
that as an SLI of `1` (all good), so a quiet night never looks like an
outage, and you never divide by zero.

### SLO target: written as a fraction

The *SLO target* is the SLI you promise over a window, often 30 days. In
this exercise a target is always a **fraction** below 1: 99.9% is written
`0.999`, 99% is `0.99`. Never `99.9`.

The share you may get wrong is `1 - target`, so `0.001` for 99.9%. A
`float64` can't hold most decimals exactly:

```go
target := 0.999
1 - target // => 0.0010000000000000009, not quite 0.001
```

That tiny error is normal, so the tests accept any answer **within 0.001**
of the expected one, and the alert examples never sit exactly on a
threshold.

### Error budget in requests

The *error budget* is how many requests may fail in the window without
breaking the SLO:

```text
allowed failures = total requests × (1 - target)
```

For a 99% target and 20,000 requests, 200 may fail. *Budget remaining* says
how much of that allowance is left, as a fraction:

| Failed so far | Remaining |
|---|---|
| 0 of 200 allowed | 1 (untouched) |
| 50 of 200 allowed | 0.75 |
| 200 of 200 allowed | 0 (spent) |
| 300 of 200 allowed | -0.5 (blown by half) |

A blown budget is **negative**, not clamped to 0, so you can see by how much.

### Burn rate: how fast you spend

The *burn rate* compares the error rate you see with the error rate the SLO
allows:

```text
error rate = 1 - SLI
burn rate  = error rate ÷ (1 - target)
```

A burn rate of **1** spends the budget exactly on schedule: it runs out at
the very end of the window. Above 1, it runs out early. For a 30-day window,
the budget lasts 30 days ÷ burn rate:

| Burn rate | Budget gone in |
|---|---|
| 1 | 30 days |
| 2 | 15 days |
| 14.4 | about 2 days (50 hours) |
| 36 | 20 hours |

```go
// a 99.5% target allows 0.5% errors; 2% errors burns 4x too fast
target := 0.995
errorRate := 0.02
errorRate / (1 - target) // => 3.9999999999999964, about 4
```

### Paging on two windows

When should on-call get paged? The Google SRE workbook suggests: page when
the burn rate is **above 14.4** over the last **1 hour**. At that speed, one
hour eats 2% of a 30-day budget.

One window alone has problems:

| Window | Problem |
|---|---|
| Only the long one (1h) | After you fix the bug, the bad hour keeps the alert firing for up to an hour. |
| Only the short one (5m) | A 2-minute blip pages someone at 3 a.m. for nothing. |

So check **both**. A *multi-window alert* pages only when the long window
**and** the short window (5 minutes, 1/12 of the long one) burn faster than
the threshold. The long window proves it's serious; the short window proves
it's still happening.

> **Why is it written like this?** Writing targets as fractions is the
> author's choice: it keeps every formula free of `/ 100`. Strictly "greater
> than" the threshold follows the workbook's alert rules. Counting good and
> total (instead of bad) is the usual convention in SLO tools, because
> "good" is what the SLO promises.

## Your task

You're on call for the checkout API, with a 99.9% SLO. Your monitoring gives
you counts of good and total requests. Write the helpers your alerting rules
will use.

### 1. Measure the SLI

**Goal:** make `SLI(good, total)` return the fraction of requests that were good, and `1` when `total` is 0.

```go
SLI(999, 1000)  // => 0.999
SLI(4950, 5000) // => 0.99
SLI(0, 10)      // => 0
SLI(0, 0)       // => 1 (no requests: nothing failed)
```

**Think it through:**
- `good` and `total` are `int`s. What happens to `999 / 1000` in integer division?
- Which case must you handle before dividing?

### 2. Count the error budget

**Goal:** make `AllowedFailures(target, total)` return how many of `total` requests may fail, and `BudgetRemaining(target, good, total)` return the fraction of that budget still left.

`AllowedFailures` is `total × (1 - target)`, as a `float64` (it can be 2.5).

`BudgetRemaining` uses it: failed requests are `total - good`, and the
result is `(allowed - failed) ÷ allowed`. It is `1` when nothing failed and
negative when the budget is blown. When `total` is 0 there is no budget and
nothing spent: return `1`. (Targets are always below 1, so with any traffic
`allowed` is above 0.)

```go
AllowedFailures(0.999, 1_000_000) // => 1000
AllowedFailures(0.99, 5_000)      // => 50
AllowedFailures(0.995, 500)       // => 2.5

BudgetRemaining(0.999, 999_750, 1_000_000) // => 0.75  (250 of 1000 failed)
BudgetRemaining(0.999, 998_500, 1_000_000) // => -0.5  (1500 of 1000 failed)
BudgetRemaining(0.99, 5_000, 5_000)        // => 1
BudgetRemaining(0.999, 0, 0)               // => 1
```

**Think it through:**
- `BudgetRemaining` gets `good`, but the budget counts failures. How do you get those?
- Which case would divide by zero?

### 3. Compute the burn rate

**Goal:** make `BurnRate(target, good, total)` return the error rate divided by `1 - target`.

The error rate is `1 - SLI(good, total)`. So with no requests the burn rate
is `0`.

```go
BurnRate(0.999, 9_990, 10_000)  // => 1     (0.1% errors, exactly on schedule)
BurnRate(0.999, 9_856, 10_000)  // => 14.4  (1.44% errors)
BurnRate(0.99, 900, 1_000)      // => 10
BurnRate(0.999, 10_000, 10_000) // => 0
BurnRate(0.999, 0, 0)           // => 0
```

**Think it through:**
- Can you reuse `SLI` so the "total is 0" rule comes for free?

### 4. Decide whether to page

**Goal:** make `ShouldPage(target, threshold, long, short)` return `true` only when the burn rate of **both** windows is greater than `threshold`.

A `Window` holds the `Good` and `Total` request counts seen over one time
window (for example the last hour, or the last 5 minutes). Equal to the
threshold is not enough: it must be strictly greater.

```go
target, threshold := 0.999, 14.4
ShouldPage(target, threshold, Window{9_800, 10_000}, Window{780, 800}) // => true  (burn 20 and 25)
ShouldPage(target, threshold, Window{9_800, 10_000}, Window{800, 800}) // => false (fixed: short window is clean)
ShouldPage(target, threshold, Window{9_990, 10_000}, Window{700, 800}) // => false (short blip: long burn is only 1)
ShouldPage(target, threshold, Window{9_800, 10_000}, Window{0, 0})     // => false (no traffic in the last 5m)
```

**Think it through:**
- `BurnRate` wants `good` and `total`. Which fields of a `Window` are those?
- "Both" is which logical operator?

## Learn more

- [Google SRE book: Service Level Objectives](https://sre.google/sre-book/service-level-objectives/)
- [Google SRE workbook: Alerting on SLOs](https://sre.google/workbook/alerting-on-slos/)
- [Google SRE workbook: Implementing SLOs](https://sre.google/workbook/implementing-slos/)
