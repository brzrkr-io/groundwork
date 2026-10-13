# For loops: plan a rollout

Rolling out to a whole fleet at once is risky, so deploy tools update hosts in
batches and retry failures with capped backoff. You'll write both plans.

## Learn

### Recap

`for i := 1; i <= n; i += step { }` counts in steps of `step`. A
condition-only loop `for x < limit { }` runs while the condition holds. Build
text with a `strings.Builder`: `b.WriteString(...)` adds a piece and
`b.String()` returns the result. Add a separator before every item except the
first. The built-in `min(a, b)` returns the smaller of two numbers:
`min(9, 7)` is `7`.

## Your task

### 1. Plan the rollout batches

Write `RolloutPlan(hosts, batchSize int) string`. Hosts are numbered from 1 to
`hosts`. Split them into batches of `batchSize` and describe each batch as
`first-last`, joined with commas. The last batch may be smaller; it still uses
`first-last`, even with one host. Zero hosts gives `""`. `batchSize` is
always at least 1. You'll need `strconv.Itoa` and `strings.Builder`, so add
both packages to the imports below the `package` line:

```go
import (
	"strconv"
	"strings"
)
```

```go
RolloutPlan(7, 3)  // => "1-3,4-6,7-7"
RolloutPlan(4, 2)  // => "1-2,3-4"
RolloutPlan(2, 5)  // => "1-2"
RolloutPlan(0, 3)  // => ""
```

### 2. Plan the retry delays

Write `RetrySchedule(attempts, baseSeconds, maxSeconds int) string`. The
first delay is `baseSeconds`, and each next delay doubles, but no delay is
ever larger than `maxSeconds`. Write each delay with an `s` suffix, joined by
single spaces. Zero attempts gives `""`.

```go
RetrySchedule(5, 1, 8)   // => "1s 2s 4s 8s 8s"
RetrySchedule(4, 5, 12)  // => "5s 10s 12s 12s"
RetrySchedule(0, 1, 8)   // => ""
```

## Learn more

- [Go by Example: For](https://gobyexample.com/for)
- [Go spec: Min and max](https://go.dev/ref/spec#Min_and_max)
- [Kubernetes: Rolling update deployments](https://kubernetes.io/docs/concepts/workloads/controllers/deployment/#rolling-update-deployment)
