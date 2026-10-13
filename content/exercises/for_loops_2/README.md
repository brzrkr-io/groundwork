# For loops: fix the off-by-one bugs

Most loop bugs are *off-by-one* errors: starting at the wrong number,
stopping one round too early or too late, or adding one separator too many.
This exercise has three loops with exactly those bugs.

**New here:** off-by-one bugs, separators between items; practice.

## Learn

### Recap

| Loop | Counts / runs |
|---|---|
| `for i := 1; i <= n; i++ { }` | `i` = 1, 2, ..., `n` |
| `for i := range n { }` | `i` = 0, 1, ..., `n-1` |
| `for x > 0 { }` | while the condition is true |

`continue` skips to the next round; `break` leaves the loop for good. A
`strings.Builder` collects text with `b.WriteString(...)`, and `b.String()`
returns the result. `strconv.Itoa(n)` turns a number into text.

### Where do off-by-one bugs hide?

- **The start:** `range n` starts at `0`, not `1`.
- **The condition:** `<` and `<=` differ by exactly one round.
- **The separator:** a comma after every item leaves one extra at the end.

### Separators

To put a separator *between* items, write it *before* every item except the
first:

```go
var b strings.Builder
for n := 10; n <= 30; n += 10 {
	if n > 10 {
		b.WriteString(" | ")
	}
	b.WriteString(strconv.Itoa(n))
}
// b.String() is "10 | 20 | 30"
```

> **Why is it written like this?** Go has no rule here; it's a common
> pattern. The first round is easy to detect (the counter is at its start
> value), while the last round is often harder to know in advance.

## Your task

You're generating config for a small web fleet. Each function is already
written but buggy. Run the tests, compare *got* with *want*, find the bug,
fix it. The `TODO` comment sits just above the buggy code.

### 1. Fix the host names

**Goal:** make `HostNames(prefix, count)` return `count` host names numbered from 1, joined with commas.

Each name is the `prefix` parameter, a fixed `-`, then the number. There's no
comma at the end. Zero hosts gives `""`. The starter has two bugs.

```go
HostNames("web", 3)   // => "web-1,web-2,web-3"
HostNames("cache", 2) // => "cache-1,cache-2"
HostNames("db", 1)    // => "db-1"
HostNames("web", 0)   // => ""
```

**Think it through:**
- What number does the starter's first name get? Which number does `range` start at?
- Where does the starter write the comma: before or after each name?

### 2. Skip the reserved port

**Goal:** make `PortList(first, last, reserved)` list every port from `first` to `last`, leaving out only `reserved`.

Both `first` and `last` are included, joined with commas. The starter stops
at the reserved port instead of skipping it. You can assume `first` is not
the reserved port. If `reserved` is outside the range, nothing is left out.

```go
PortList(8000, 8003, 8002) // => "8000,8001,8003"
PortList(80, 82, 81)       // => "80,82"
PortList(9000, 9002, 7000) // => "9000,9001,9002"
PortList(5000, 5000, 1)    // => "5000"
```

**Think it through:**
- What happens to the ports *after* the reserved one in the starter?
- Which keyword ends the loop, and which one skips just one round?

### 3. Count the batches

**Goal:** make `BatchCount(total, batchSize)` return how many batches you need for `total` items.

Each batch holds `batchSize` items; the last batch may be smaller. The
starter forgets that last, partial batch. `batchSize` is always at least 1.

```go
BatchCount(10, 3) // => 4  (3 + 3 + 3 + 1)
BatchCount(9, 3)  // => 3
BatchCount(2, 5)  // => 1
BatchCount(0, 3)  // => 0
```

**Think it through:**
- Which test passes already (`9, 3`) and which fails (`10, 3`)? What's different about them?
- Look at the loop condition: when does it stop, and is anything left over?

## Learn more

- [A Tour of Go: For](https://go.dev/tour/flowcontrol/1)
- [Go by Example: For](https://gobyexample.com/for)
- [Go spec: For statements](https://go.dev/ref/spec#For_statements)
