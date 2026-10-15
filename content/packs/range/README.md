# Looping with range

Monitoring data comes in lists: one latency per host, one health check per
node. `for ... range` walks a slice element by element, so you don't manage an
index by hand.

**New here:** `for ... range` over slices and strings, the blank identifier `_`.

## Learn

### range over a slice

So far you looped with `for i := range len(s)` and read `s[i]`. Ranging over
the slice itself hands you the index **and** the element:

```go
zones := []string{"eu-1", "us-1", "ap-1"}
for i, zone := range zones {
	fmt.Println(i, zone)
}
// 0 eu-1
// 1 us-1
// 2 ap-1
```

Each pass, `i` is the index and `zone` is `zones[i]`. If the slice is empty or
`nil`, the body never runs.

> **Why is it written like this?** `for`, `range`, `:=` and the comma are
> required. The order is fixed: index first, element second. The names `i` and
> `zone` are your choice; `i` is a convention.

### Skip what you don't need

Go refuses to compile a variable you never use. If you only need the element,
put the *blank identifier* `_` where the index goes:

```go
longest := 0
for _, name := range zones {
	if len(name) > longest {
		longest = len(name)
	}
}
```

If you only need the index, write just one variable. Careful: one variable
means the **index**, not the element.

| Loop | You get |
|---|---|
| `for i, v := range s` | index and element |
| `for _, v := range s` | element only |
| `for i := range s` | index only |

> **Why is it written like this?** `_` is required when you skip the index,
> because the position decides the meaning. It is Go's "throw this away" name.

### The element is a copy

`v` is a copy of `s[i]`. Assigning to `v` changes only the copy. To change the
slice, assign through the index: `s[i] = ...`.

```go
sizes := []int{3, 9}
for _, v := range sizes {
	v = v * 10     // changes the copy only
	fmt.Println(v) // 30, then 90
}
// sizes is still [3 9]

for i := range sizes {
	sizes[i] = sizes[i] * 10 // changes the slice itself
}
// sizes is now [30 90]
```

### Two slices, one index

Sometimes two slices line up: `zones[i]` belongs with `owners[i]`. Range over
one and use the index to read the other:

```go
for i, zone := range zones {
	fmt.Println(zone, owners[i])
}
```

### range over a string

`range` also works on strings. It gives the byte position and the character,
called a *rune*:

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

**Goal:** make `TotalLatency(ms)` return the sum of every value in `ms`.

An empty or `nil` slice gives 0.

```go
TotalLatency([]int{120, 80, 300})
// => 500

TotalLatency(nil)
// => 0
```

**Think it through:**
- Do you need the index, the element, or both?
- `total` already exists in the starter. What happens to it each pass?

### 2. Find the slowest host

**Goal:** make `Slowest(hosts, ms)` return the host with the highest latency,
and that latency.

- `hosts[i]` has latency `ms[i]`. Both slices always have the same length.
- Latencies are never negative.
- On a tie, return the host that comes first.
- If every latency is 0, the first host still wins.
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

**Think it through:**
- `slowest` and `worst` hold the slowest host so far. When should they be replaced?
- Which comparison keeps the first host on a tie: "greater" or "greater or equal"?
- With all zeros, nothing is greater than 0. How does the first host get picked?

### 3. List the unhealthy hosts

**Goal:** make `Unhealthy(hosts, healthy)` return the hosts whose check failed,
in their original order.

- A check failed when `healthy[i]` is `false`; the host that failed is `hosts[i]`.
- Both slices have the same length.
- If every host is healthy, return a slice with no elements (`nil` is fine).

```go
Unhealthy([]string{"web-1", "web-2", "db-1"}, []bool{true, false, false})
// => ["web-2" "db-1"]

Unhealthy([]string{"web-1"}, []bool{true})
// => []
```

**Think it through:**
- Which slice do you range over, and which do you read with the index?
- How do you test that a `bool` is `false`?

## Learn more

- [A Tour of Go: Range](https://go.dev/tour/moretypes/16)
- [Go by Example: Range over Built-in Types](https://gobyexample.com/range-over-built-in-types)
- [Go spec: For statements with range clause](https://go.dev/ref/spec#For_range)
