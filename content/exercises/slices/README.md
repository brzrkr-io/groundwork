# Slices

A list of hosts in an inventory, the ports a service listens on, the lines of
a log file: most collections don't have a size you know in advance. Go uses a
*slice* for them, and you will see slices in almost every Go program.

## Learn

### A slice is a list without a fixed length

The type `[]string` (no number in the brackets) is a slice of strings. A slice
literal looks like an array literal without the length:

```go
hosts := []string{"web-1", "web-2", "web-3"}
```

Indexing and `len` work just as with arrays from *Arrays*:

```go
hosts[0]   // "web-1"
hosts[2]   // "web-3"
len(hosts) // 3
hosts[1] = "web-9"
```

Because the length is not part of the type, one function can accept slices of
any length: `func Count(hosts []string) int` works for 2 hosts or 2000.

### Out of range is a panic

Reading `hosts[3]` from a 3-element slice is a mistake the compiler can't
catch, because the length is only known when the program runs. The program
stops with a *panic*:

```text
panic: runtime error: index out of range [3] with length 3
```

Check `len` before you index if you aren't sure the element exists.

### Slicing: taking a part

`s[a:b]` gives you the elements from index `a` up to, but **not** including,
index `b`. The result has `b - a` elements. Leave out `a` to start at the
beginning, or leave out `b` to go to the end:

```go
hosts := []string{"web-1", "web-2", "web-3", "web-4"}
hosts[1:3] // ["web-2" "web-3"]
hosts[:2]  // ["web-1" "web-2"]          (the first 2)
hosts[2:]  // ["web-3" "web-4"]          (everything from index 2)
hosts[4:]  // []                         (nothing left, still fine)
hosts[:5]  // panic: slice bounds out of range [:5] with capacity 4
```

`a` and `b` may each be anything from `0` to `len(hosts)`, as long as `a` is
not bigger than `b`. Anything else panics. A handy pattern:
`hosts[len(hosts)-n:]` is the last `n` elements.

A slice made by slicing shares its elements with the original; it does not
copy them. That matters later, in *Slice gotchas: shared memory*. For now,
just don't change elements of a slice you were handed.

### make: a slice of a given length

`make([]T, n)` creates a new slice of `n` elements, all set to the zero value.
Then you fill it by index:

```go
ports := make([]int, 3) // [0 0 0]
for i := range len(ports) {
	ports[i] = 9000 + i
}
// ports is [9000 9001 9002]
```

### nil and empty slices

A slice variable that was declared but never set is `nil`: it has no elements
at all. `len` of a `nil` slice is `0`, and slicing it with `[:0]` or `[0:]`
works fine.

```go
var none []string     // nil, len 0
empty := []string{}   // not nil, but also len 0
none == nil           // true
```

For most code, `nil` and empty behave the same: both have length `0`. The
tests in this exercise only check the length and the elements, never whether a
slice is `nil`.

## Your task

Your deploy tool works on a list of hosts. Write three small slice helpers for
it. Don't use `append` yet (it comes in *append and copy*).

### 1. The first n hosts

Change `FirstN(hosts []string, n int) []string` so it returns the first `n`
hosts. If `n` is bigger than the number of hosts, return all of them. `n` is
never negative. `FirstN` of a `nil` or empty slice returns an empty result
(length `0`).

```go
FirstN([]string{"web-1", "web-2", "web-3"}, 2)
// => []string{"web-1", "web-2"}

FirstN([]string{"web-1", "web-2"}, 5)
// => []string{"web-1", "web-2"}

FirstN(nil, 3)
// => []string{} (length 0)
```

### 2. The last n hosts

Change `LastN(hosts []string, n int) []string` so it returns the last `n`
hosts, in their original order. The same rules as task 1 apply: too big an
`n` means all hosts, `n` is never negative, and `nil` gives an empty result.

```go
LastN([]string{"web-1", "web-2", "web-3"}, 2)
// => []string{"web-2", "web-3"}

LastN([]string{"web-1"}, 4)
// => []string{"web-1"}
```

### 3. A range of ports

Change `PortRange(start, count int) []int` so it returns `count` consecutive
port numbers beginning at `start`. Use `make`. A `count` of `0` returns an
empty slice.

```go
PortRange(8080, 3)
// => []int{8080, 8081, 8082}

PortRange(9100, 0)
// => []int{}
```

## Learn more

- [A Tour of Go: Slices](https://go.dev/tour/moretypes/7)
- [Go by Example: Slices](https://gobyexample.com/slices)
- [Go blog: Go Slices: usage and internals](https://go.dev/blog/slices-intro)
