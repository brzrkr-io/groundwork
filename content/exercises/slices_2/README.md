# Slices: fix the inventory script

Your team's inventory script reads a list of hosts and picks parts of it. It
was written in a hurry: it skips a host, keeps a host it should drop, and
crashes on an empty inventory. Time to fix it.

**New here:** reading a panic message, off-by-one bugs. The rest is practice
from *Slices*.

## Learn

### Recap: slice bounds

`s[a:b]` holds the elements from index `a` up to, but **not** including, `b`.

| Expression | Means |
| --- | --- |
| `s[a:b]` | from `a`, stop just before `b` |
| `s[a:]` | from `a` to the end |
| `s[:b]` | from the start, stop just before `b` |

```go
pods := []string{"p0", "p1", "p2", "p3"}
pods[1:3] // ["p1" "p2"]   (index 3 is left out)
pods[2:]  // ["p2" "p3"]
pods[:2]  // ["p0" "p1"]
```

> **Why is it written like this?** Go requires the colon and the square
> brackets. Leaving a bound out is allowed by Go and is the convention when
> you mean "the start" or "the end".

### Off-by-one bugs

Most slicing bugs are *off-by-one*: a bound that is one too big or one too
small. When in doubt, write the indexes under the values:

```text
lines:  "name"  "web-1"  "web-2"  "web-3"
index:    0        1        2        3       len(lines) = 4
```

The valid indexes run from `0` to `len(s)-1`. So the last element is
`s[len(s)-1]`, and an empty slice has **no** valid index at all.

### Reading a panic

Using an index that doesn't exist makes the program *panic*: it stops and
prints a message plus a *stack trace* (the functions that were running).

```text
panic: runtime error: index out of range [-1] [recovered, repanicked]
...
...LastHost(...)
	.../exercise.go:21
```

Read the first line for **what** went wrong. Then look for the line naming
your function and `exercise.go`: that is **where** it happened.

> **Why is it written like this?** Go requires every index to be checked at
> run time; a bad one panics instead of reading random memory. Guarding with
> an `if` before indexing is your choice of fix, and the usual one.

## Your task

Fix the three bugs in `exercise.go`. Each fix changes one or two lines. Don't
rewrite the functions.

### 1. Skip the header, keep every host

**Goal:** make `SkipHeader` return every line except the first one.

The first line is a header (`"name"`); every line after it is a host. Right
now the last host goes missing. `lines` always has the header plus at least
one host.

```go
SkipHeader([]string{"name", "web-1", "web-2", "web-3"})
// => []string{"web-1", "web-2", "web-3"}
```

**Think it through:**
- Which index does the first host have? Which bound in the line controls that?
- The end bound is left out of the result. Is the last host being left out?

### 2. Drain one host

**Goal:** make `Drain` return the hosts before index `i` and the hosts after
it, leaving `hosts[i]` out of both.

Right now `hosts[i]` shows up at the start of `after`. `i` is always a valid
index. When nothing is before or after `i`, that slice is empty.

```go
before, after := Drain([]string{"web-1", "web-2", "web-3", "web-4"}, 1)
// before => []string{"web-1"}
// after  => []string{"web-3", "web-4"}
```

**Think it through:**
- The `before` line is already right. Only the `after` line needs a change.
- At which index does the first host *after* `hosts[i]` sit?

### 3. Don't crash on an empty inventory

**Goal:** make `LastHost` return the last host, or `""` when `hosts` is empty
or `nil`.

Right now it panics on an empty list.

```go
LastHost([]string{"web-1", "web-2"}) // => "web-2"
LastHost([]string{})                 // => ""
LastHost(nil)                        // => ""
```

**Think it through:**
- What is `len(hosts)-1` when `hosts` is empty? Is that a valid index?
- Check for the empty case *before* the line that indexes. What is `len` of a
  `nil` slice?

## Learn more

- [A Tour of Go: Slices](https://go.dev/tour/moretypes/7)
- [Go by Example: Slices](https://gobyexample.com/slices)
- [Go spec: Slice expressions](https://go.dev/ref/spec#Slice_expressions)
