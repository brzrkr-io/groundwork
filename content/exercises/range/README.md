# Looping with range

Monitoring data comes in lists: one latency per host, one health check per
node. `for ... range` walks a slice element by element without you managing an
index by hand.

## Learn

### range over a slice

So far you looped with `for i := range len(s)` and read `s[i]`. Ranging over
the slice itself gives you both the index and the element:

```go
hosts := []string{"web-1", "web-2", "db-1"}
for i, host := range hosts {
	fmt.Println(i, host)
}
// 0 web-1
// 1 web-2
// 2 db-1
```

On each pass, `i` is the index and `host` is `hosts[i]`. You pick the names.
If the slice is empty or `nil`, the loop body never runs.

### Skip what you don't need

Go refuses to compile a variable you never use. If you only need the element,
put the *blank identifier* `_` where the index would go:

```go
total := 0
for _, ms := range latencies {
	total += ms
}
```

If you only need the index, leave out the second variable:

```go
for i := range hosts {
	fmt.Println(i) // 0, 1, 2
}
```

Careful: with one variable you get the **index**, not the element.

| Loop | You get |
|---|---|
| `for i, v := range s` | index and element |
| `for _, v := range s` | element only |
| `for i := range s` | index only |

### The element is a copy

`v` is a copy of `s[i]`. Assigning to `v` changes the copy, not the slice:

```go
ms := []int{120, 900}
for _, v := range ms {
	v = 0 // only changes the copy
}
// ms is still [120 900]

for i := range ms {
	ms[i] = 0 // changes the slice
}
// ms is now [0 0]
```

To change the slice, assign through the index: `s[i] = ...`.

### Two slices, one index

Sometimes two slices line up: `hosts[i]` has latency `ms[i]`. Range over one
and use the index to read the other:

```go
for i, host := range hosts {
	fmt.Println(host, ms[i])
}
```

### range over a string

`range` also works on strings. It gives the byte position and the character
(called a *rune*):

```go
for i, r := range "db-1" {
	fmt.Println(i, string(r))
}
// 0 d
// 1 b
// 2 -
// 3 1
```

A character in single quotes, like `'.'`, is a rune you can compare with:
`if r == '.'`. The *Runes* exercise covers the details.

## Your task

Your monitoring tool collects one latency (in milliseconds) and one health
check result per host. Write three helpers for the dashboard.

### 1. Add up the latencies

`TotalLatency(ms)` returns the sum of all values in `ms`. An empty slice gives
0.

```go
TotalLatency([]int{120, 80, 300})
// => 500

TotalLatency(nil)
// => 0
```

### 2. Find the slowest host

`Slowest(hosts, ms)` returns the host with the highest latency and that
latency. `hosts[i]` has latency `ms[i]`; both slices always have the same
length and latencies are never negative.

- If two hosts tie, return the one that comes first.
- If `hosts` is empty, return `""` and `0`.

```go
Slowest([]string{"web-1", "web-2", "db-1"}, []int{120, 450, 300})
// => "web-2", 450

Slowest([]string{"web-1", "web-2"}, []int{200, 200})
// => "web-1", 200

Slowest([]string{"web-1", "web-2"}, []int{0, 0})
// => "web-1", 0

Slowest(nil, nil)
// => "", 0
```

### 3. List the unhealthy hosts

`Unhealthy(hosts, healthy)` returns the hosts whose health check failed
(`healthy[i]` is `false`), in their original order. Both slices have the same
length. If every host is healthy, return a slice with no elements (`nil` is
fine).

```go
Unhealthy([]string{"web-1", "web-2", "db-1"}, []bool{true, false, false})
// => ["web-2" "db-1"]

Unhealthy([]string{"web-1"}, []bool{true})
// => []
```

## Learn more

- [A Tour of Go: Range](https://go.dev/tour/moretypes/16)
- [Go by Example: Range over Built-in Types](https://gobyexample.com/range-over-built-in-types)
- [Go spec: For statements with range clause](https://go.dev/ref/spec#For_range)
