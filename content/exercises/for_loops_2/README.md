# For loops: fix the off-by-one bugs

Most loop bugs are *off-by-one* errors: starting at the wrong number, stopping
one round too early or too late, or adding one separator too many. This
exercise has three loops with exactly those bugs.

## Learn

### Recap

A three-part loop `for i := 1; i <= n; i++ { }` counts from 1 to `n`;
`for i := range n` counts from 0 to `n-1`. A condition-only loop
`for x > 0 { }` runs while the condition is true. `continue` skips to the next
round; `break` leaves the loop. A `strings.Builder` collects pieces of text
with `b.WriteString(...)`, and `b.String()` returns the result.

### Separators

To join items with commas, add the comma *before* every item except the first:

```go
var b strings.Builder
for i := range 3 {
	if i > 0 {
		b.WriteString(",")
	}
	b.WriteString(strconv.Itoa(i))
}
// b.String() is "0,1,2"
```

## Your task

You're generating config for a small web fleet. Each function below is
already written but buggy. Read the failing test, find the bug, fix it.

### 1. Fix the host names

`HostNames(prefix string, count int) string` must return `count` names
numbered from 1, joined with commas and no trailing comma. Zero hosts gives
`""`.

```go
HostNames("web", 3) // => "web-1,web-2,web-3"
HostNames("db", 1)  // => "db-1"
HostNames("web", 0) // => ""
```

### 2. Skip the reserved port

`PortList(first, last, reserved int) string` must list every port from
`first` to `last` (both included), joined with commas, but leave out
`reserved`. The starter stops at the reserved port instead of skipping it.
Use `continue`. You can assume `first` is not the reserved port.

```go
PortList(8000, 8003, 8002) // => "8000,8001,8003"
PortList(9000, 9002, 7000) // => "9000,9001,9002"
```

### 3. Count the batches

`BatchCount(total, batchSize int) int` returns how many batches you need to
process `total` items, `batchSize` at a time. The last batch may be smaller.
The starter forgets that last, partial batch. `batchSize` is always at least 1.

```go
BatchCount(10, 3) // => 4  (3 + 3 + 3 + 1)
BatchCount(9, 3)  // => 3
BatchCount(0, 3)  // => 0
```

## Learn more

- [A Tour of Go: For](https://go.dev/tour/flowcontrol/1)
- [Go by Example: For](https://gobyexample.com/for)
- [Go spec: For statements](https://go.dev/ref/spec#For_statements)
