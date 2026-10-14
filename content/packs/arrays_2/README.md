# Arrays: fix the hourly histogram

A load balancer keeps a *histogram* of requests: one counter for each of the
24 hours of the day. A teammate wrote three helpers for it, and each one has a
bug.

**New here:** keyed array literals; reading and fixing someone else's code.

## Learn

### Recap from *Arrays*

- `[24]int` holds exactly 24 ints, indexed `0` to `23`.
- `for i := range len(hours)` visits every index, from `0` to the last.
- After `b := a`, changing `b` does not change `a`. They are separate copies.

### Reading someone else's code

Each function has one wrong line, marked with a `TODO`. Read the code, compare
it with the description below, and find what it does differently. Each fix is
small. Don't rewrite the functions from scratch.

### Keyed array literals

Writing out 24 numbers is tiring, so the tests use a *keyed* array literal:
`index: value` inside the braces. Every index you leave out gets the zero
value.

```go
cpu := [8]int{2: 75, 5: 30}
// cpu[2] is 75, cpu[5] is 30, the other 6 are 0
```

> **Why is it written like this?** The `index: value` form is Go syntax; the
> index must be a constant. Listing keys in order is convention, not required.

## Your task

Fix the three bugs in `exercise.go`.

### 1. Count every hour

**Goal:** make `TotalRequests` add up all 24 counters, including hour `0`.

Right now the total is wrong whenever hour `0` (midnight) had traffic.

```go
TotalRequests([24]int{0: 5, 13: 20})
// => 25
```

**Think it through:**
- Look at the loop line. Which index does it start at?
- Which index is the first element of an array?

### 2. Find the peak hour

**Goal:** make `PeakHour` return the hour (an index, `0` to `23`) with the
most requests.

If several hours tie for the most, return the earliest. An all-zero day
returns `0`.

```go
PeakHour([24]int{3: 100, 14: 900, 20: 100})
// => 14

PeakHour([24]int{9: 300, 12: 10, 17: 300})
// => 9 (hours 9 and 17 tie, 9 is earlier)
```

**Think it through:**
- `best` holds an *hour* (an index). `hours[i]` is a *request count*. The `if`
  compares a count with an hour. What should it compare with instead?
- How do you get the request count of the hour stored in `best`?
- Keep `>` (not `>=`), so a later tie does not replace the earlier hour.

### 3. Clear one bad hour

**Goal:** make `ClearHour` return a copy of `hours` with index `hour` set to
`0`.

Every other hour keeps its value. `hour` is always between `0` and `23`.

```go
ClearHour([24]int{3: 99999, 4: 40}, 3)
// => [24]int{4: 40}
```

**Think it through:**
- The function makes a copy called `cleaned` and returns it. Which array does
  the next line change?
- Remember: changing one copy never changes the other.

## Learn more

- [A Tour of Go: Arrays](https://go.dev/tour/moretypes/6)
- [Go spec: Composite literals](https://go.dev/ref/spec#Composite_literals)
- [Go spec: Assignment statements](https://go.dev/ref/spec#Assignment_statements)
