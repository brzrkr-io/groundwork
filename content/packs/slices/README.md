# Slices

A list of hosts in an inventory, the ports a service listens on, the lines of
a log file: most collections don't have a size you know in advance. Go uses a
*slice* for them, and you will see slices in almost every Go program.

**New here:** slice types like `[]string`, slicing with `s[a:b]`, `make`,
`nil` slices.

## Learn

### A slice is a list without a fixed length

`[]string` (no number in the brackets) is a slice of strings. A slice literal
looks like an array literal without the length:

```go
users := []string{"ana", "ben", "cho"}
```

Indexing and `len` work as with *Arrays*:

```go
users[0]   // "ana"
len(users) // 3
users[1] = "bo"
```

Because the length is not part of the type, one function can take slices of
any length: `func Count(users []string) int` works for 2 users or 2000.

> **Why is it written like this?** Empty brackets `[]T` are how Go spells a
> slice type; a number inside would make it an array. That's required syntax.

### Out of range is a panic

Reading `users[3]` from a 3-element slice can't be caught by the compiler,
because the length is only known when the program runs. The program stops
with a *panic*:

```text
panic: runtime error: index out of range [3] with length 3
```

Check `len` first if you aren't sure the element exists.

### Slicing: taking a part

`s[a:b]` gives the elements from index `a` up to, but **not** including,
index `b`. The result has `b - a` elements. Leave out `a` to start at the
beginning; leave out `b` to go to the end.

```go
logs := []string{"l0", "l1", "l2", "l3", "l4"}
logs[1:4] // ["l1" "l2" "l3"]
logs[3:]  // ["l3" "l4"]   (from index 3 to the end)
logs[:1]  // ["l0"]        (up to, not including, index 1)
logs[5:]  // []            (nothing left, still fine)
logs[:6]  // panic: slice bounds out of range [:6] with capacity 5
```

| Expression | Means |
| --- | --- |
| `s[a:b]` | from `a` up to (not incl.) `b` |
| `s[a:]` | from `a` to the end |
| `s[:b]` | from the start up to (not incl.) `b` |

`a` and `b` can be anything from `0` to `len(s)`, with `a` not bigger than
`b`. Anything else panics. Both may be expressions, like `s[i+1:]`.

> **Why is it written like this?** The colon and square brackets are required
> syntax. "Up to but not including" is a Go rule; it makes `s[:k]` have
> exactly `k` elements.

A slice made by slicing shares its elements with the original; it does not
copy them. That matters later, in *Slice gotchas: shared memory*. For now,
don't change elements of a slice you were handed.

### make: a slice of a given length

`make([]T, n)` creates a new slice of `n` elements, all set to the zero value.
Then you fill it by index:

```go
squares := make([]int, 4) // [0 0 0 0]
for i := range len(squares) {
	squares[i] = i * i
}
// squares is [0 1 4 9]
```

> **Why is it written like this?** `make` is built into Go; the first
> argument is the type, the second the length. `n` may be a variable or `0`.

### nil and empty slices

A slice variable declared but never set is `nil`. `len` of a `nil` slice is
`0`, and slicing it with `[:0]` or `[0:]` works fine.

```go
var none []string   // nil, len 0
empty := []string{} // not nil, also len 0
```

For most code they behave the same. The tests only check length and
elements, never whether a slice is `nil`.

## Your task

Your deploy tool works on a list of hosts. Write three small slice helpers.
Don't use `append` yet (it comes in *append and copy*).

### 1. The first n hosts

**Goal:** make `FirstN` return the first `n` elements of `hosts`.

If `n` is bigger than the number of hosts, return all of them (don't panic).
`n` is never negative. `n` of `0`, or a `nil`/empty `hosts`, returns an empty
result (length `0`).

```go
FirstN([]string{"web-1", "web-2", "web-3"}, 2)
// => []string{"web-1", "web-2"}

FirstN([]string{"web-1", "web-2"}, 5)
// => []string{"web-1", "web-2"}

FirstN(nil, 3)
// => []string{} (length 0)
```

**Think it through:**
- Which slicing form from the table means "from the start"?
- What should happen to `n` when it is bigger than `len(hosts)`? You may
  change a parameter inside the function.

### 2. The last n hosts

**Goal:** make `LastN` return the last `n` elements of `hosts`, in their
original order.

Same rules as task 1: too big an `n` means all hosts, `n` is never negative,
and an `n` of `0` or a `nil` `hosts` gives an empty result.

```go
LastN([]string{"web-1", "web-2", "web-3"}, 2)
// => []string{"web-2", "web-3"}

LastN([]string{"web-1"}, 4)
// => []string{"web-1"}
```

**Think it through:**
- In the example, the result starts at index 1 of 3 hosts. How do you get
  that 1 from the length and `n`?
- Which slicing form means "to the end"?
- Guard against a big `n` the same way as in task 1.

### 3. A range of ports

**Goal:** make `PortRange` return `count` consecutive port numbers, starting
at `start`.

Use `make`. A `count` of `0` returns an empty slice.

```go
PortRange(8080, 3)
// => []int{8080, 8081, 8082}

PortRange(9100, 0)
// => []int{}
```

**Think it through:**
- How long must the slice be? Which parameter tells you?
- The element at index `i` is how far past `start`?

## Learn more

- [A Tour of Go: Slices](https://go.dev/tour/moretypes/7)
- [Go by Example: Slices](https://gobyexample.com/slices)
- [Go blog: Go Slices: usage and internals](https://go.dev/blog/slices-intro)
