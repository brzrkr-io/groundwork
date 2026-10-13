# Arrays: fix the hourly histogram

A load balancer keeps a *histogram* of requests: one counter for each of the
24 hours of the day. A teammate wrote three helpers for it, and each one has a
bug.

## Learn

Recap from *Arrays*: a `[24]int` holds exactly 24 ints, indexed `0` to `23`.
Loop over it with `for i := range len(hours)`, so you never miss the first or
the last element. Arrays are copied on assignment: after `b := a`, changing
`b` does not change `a`.

Reading someone else's code is a big part of ops work. For each function, read
the code, compare it with the description below, and find the line that does
something different. The fix is small every time.

### New: keyed array literals

Writing out 24 numbers is tiring, so the tests use a *keyed* array literal.
You put `index: value` inside the braces, and every index you leave out gets
the zero value:

```go
hours := [24]int{3: 40, 14: 900}
// hours[3] is 40, hours[14] is 900, the other 22 hours are 0
```

## Your task

Fix the three bugs in `exercise.go`. Don't rewrite the functions from
scratch; find and fix the broken line.

### 1. Count every hour

`TotalRequests(hours [24]int) int` should return the sum of all 24 counters.
Right now it gets the wrong total whenever hour `0` (midnight) had traffic.
Rewrite the loop line as a `range` loop that visits every hour; your editor
suggests the same if you write `for i := 0; i < len(hours); i++`.

```go
TotalRequests([24]int{0: 5, 13: 20})
// => 25
```

### 2. Find the peak hour

`PeakHour(hours [24]int) int` should return the hour (the index, `0` to `23`)
with the most requests. If several hours tie for the most, return the earliest
of them. An all-zero day returns `0`.

```go
PeakHour([24]int{3: 100, 14: 900, 20: 100})
// => 14

PeakHour([24]int{9: 300, 12: 10, 17: 300})
// => 9 (hours 9 and 17 tie, 9 is earlier)
```

### 3. Clear one bad hour

During a monitoring outage, one hour recorded garbage.
`ClearHour(hours [24]int, hour int) [24]int` should return a copy of the
histogram with that hour set to `0`. Every other hour keeps its value. `hour`
is always between `0` and `23`.

```go
ClearHour([24]int{3: 99999, 4: 40}, 3)
// => [24]int{4: 40}
```

## Learn more

- [A Tour of Go: Arrays](https://go.dev/tour/moretypes/6)
- [Go spec: Assignment statements](https://go.dev/ref/spec#Assignment_statements)
