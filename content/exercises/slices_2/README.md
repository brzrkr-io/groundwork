# Slices: fix the inventory script

Your team's inventory script reads a list of hosts and picks parts of it. It
was written in a hurry: it skips a host, keeps a host it should drop, and
crashes on an empty inventory. Time to fix it.

## Learn

Recap from *Slices*: `s[a:b]` holds the elements from index `a` up to, but not
including, `b`. `s[a:]` runs to the end and `s[:b]` starts at the beginning.
The valid indexes of `s` are `0` to `len(s)-1`, so the last element is
`s[len(s)-1]`. Indexing or slicing past the end makes the program *panic*.

When a test panics, Go prints the panic message and a *stack trace* (the list
of functions that were running), then stops, so tests after it don't run. The
first line tells you what went wrong. Further down, look for the line that
names your function and `exercise.go`: it shows where the panic happened.

```text
panic: runtime error: index out of range [-1] [recovered, repanicked]
...
...LastHost(...)
	.../exercise.go:21
...
```

Most slicing bugs are *off-by-one* errors: a bound that is one too big or one
too small. Draw the indexes on paper when in doubt:

```text
lines:  "name"  "web-1"  "web-2"  "web-3"
index:    0        1        2        3       len(lines) = 4
```

## Your task

Fix the three bugs in `exercise.go`. Each fix is a change to one or two lines.

### 1. Skip the header, keep every host

The inventory file's first line is a header (`"name"`), and every line after it
is a host. `SkipHeader(lines []string) []string` should return all the lines
except the header. Right now it also drops the last host. `lines` always has
the header plus at least one host.

```go
SkipHeader([]string{"name", "web-1", "web-2", "web-3"})
// => []string{"web-1", "web-2", "web-3"}
```

### 2. Drain one host

To patch a host, you take it out of the rotation.
`Drain(hosts []string, i int) (before, after []string)` should return the
hosts before index `i` and the hosts after it. The drained host `hosts[i]` is
in neither slice. Right now it shows up in `after`. `i` is always a valid
index.

```go
before, after := Drain([]string{"web-1", "web-2", "web-3", "web-4"}, 1)
// before => []string{"web-1"}
// after  => []string{"web-3", "web-4"}
```

### 3. Don't crash on an empty inventory

`LastHost(hosts []string) string` should return the last host, or `""` when
`hosts` is empty. Right now it panics when `hosts` is empty.

```go
LastHost([]string{"web-1", "web-2"}) // => "web-2"
LastHost([]string{})                 // => ""
LastHost(nil)                        // => ""
```

## Learn more

- [A Tour of Go: Slices](https://go.dev/tour/moretypes/7)
- [Go by Example: Slices](https://gobyexample.com/slices)
- [Go spec: Slice expressions](https://go.dev/ref/spec#Slice_expressions)
