# For loops: plan a rollout

Rolling out to a whole fleet at once is risky, so deploy tools update hosts in
batches and retry failures with capped backoff. You'll write both plans.

**New here:** counting in steps (`i += step`), the built-in `min`; practice.

## Learn

### Recap

- A condition-only loop `for x < limit { }` runs while the condition holds.
- `strings.Builder` builds text: `b.WriteString(...)` adds a piece, `b.String()` returns it.
- To separate items, write the separator before every item except the first.
- `strconv.Itoa(n)` turns a number into text.

### Counting in steps

The post part of a three-part loop doesn't have to be `i++`. Any change works:

```go
for t := 0; t < 60; t += 15 {
	fmt.Println(t) // prints 0, 15, 30, 45
}
```

> **Why is it written like this?** Go requires the post part to be a single
> statement, such as `t++` or `t += 15`. The step size is your choice.

### The smaller of two numbers

The built-in `min(a, b)` returns the smaller number (Go 1.21 and newer). It's
handy for a *cap*, a limit a value must never go over:

```go
min(9, 7)       // => 7
min(3, 7)       // => 3
seats := min(guests, 50) // never more than 50
```

> **Why is it written like this?** `min` is *built in*: no import, no
> package name in front, like `len`. There's also `max`.

### Adding imports

This starter has no imports. You need two packages, so use the grouped form,
one package per line inside `( )`:

```go
import (
	"fmt"
	"os"
)
```

> **Why is it written like this?** Go requires imports right after the
> `package` line, and rejects packages you import but don't use. gofmt sorts
> the list alphabetically.

## Your task

The starter functions return a placeholder; replace each one. Both need
`strconv` and `strings`, so add those two packages to the imports below the
`package` line.

### 1. Plan the rollout batches

**Goal:** make `RolloutPlan(hosts, batchSize)` describe the batches as `first-last` ranges joined with commas.

- Hosts are numbered from `1` to `hosts`.
- Each batch holds `batchSize` hosts; the last one may be smaller.
- Each batch is written as its first host, a fixed `-`, and its last host, even when they're the same (`7-7`).
- Batches are joined with commas, no comma at the end.
- Zero hosts gives `""`. `batchSize` is always at least 1.

```go
RolloutPlan(7, 3)  // => "1-3,4-6,7-7"
RolloutPlan(10, 4) // => "1-4,5-8,9-10"
RolloutPlan(4, 2)  // => "1-2,3-4"
RolloutPlan(2, 5)  // => "1-2"
RolloutPlan(0, 3)  // => ""
```

**Think it through:**
- Which numbers does each batch *start* at? How far apart are they?
- How do you get a batch's last host, and what must it never be bigger than?

### 2. Plan the retry delays

**Goal:** make `RetrySchedule(attempts, baseSeconds, maxSeconds)` list `attempts` delays like `"1s 2s 4s 8s 8s"`.

- The first delay is `baseSeconds`; each next delay doubles.
- No delay written is ever larger than `maxSeconds`.
- Each delay is the number followed by the fixed text `s`.
- Delays are separated by single spaces, no space at the end.
- Zero attempts gives `""`.

```go
RetrySchedule(5, 1, 8)   // => "1s 2s 4s 8s 8s"
RetrySchedule(4, 5, 12)  // => "5s 10s 12s 12s"
RetrySchedule(2, 3, 100) // => "3s 6s"
RetrySchedule(0, 1, 8)   // => ""
```

**Think it through:**
- Which value doubles every round, and which value is the cap?
- What comes before each delay except the first, and what comes after every one?

## Learn more

- [Go by Example: For](https://gobyexample.com/for)
- [Go spec: Min and max](https://go.dev/ref/spec#Min_and_max)
- [Kubernetes: Rolling update deployments](https://kubernetes.io/docs/concepts/workloads/controllers/deployment/#rolling-update-deployment)
