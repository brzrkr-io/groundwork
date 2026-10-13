## Task 1: The first n hosts

Slicing can take the first part of a slice for you. The catch: slicing past the end panics, so make sure `n` is never bigger than `len(hosts)`.

### Closer

`hosts[:n]` is the first `n` hosts. Shrink `n` first with an `if`:

```go
if n > len(hosts) {
	n = len(hosts)
}
```

### Answer

```go
if n > len(hosts) {
	n = len(hosts)
}
return hosts[:n]
```

A `nil` slice has length `0`, so `n` becomes `0` and `hosts[:0]` is an empty result.

## Task 2: The last n hosts

This is like task 1, but you start partway through the slice instead of at the beginning. Work out the index where the last `n` hosts begin.

### Closer

The last `n` elements start at index `len(hosts)-n`, so the slice is `hosts[len(hosts)-n:]`. Shrink `n` the same way as in task 1, or the start index goes negative and panics.

### Answer

```go
if n > len(hosts) {
	n = len(hosts)
}
return hosts[len(hosts)-n:]
```

With `n` capped, `len(hosts)-n` is between `0` and `len(hosts)`, which is always a valid start.

## Task 3: A range of ports

Create a slice of the right length with `make`, then fill each element by index. The port at index `i` is `start` plus `i`.

### Closer

`make([]int, count)` gives you `count` zeros. Loop with `for i := range count` and set `ports[i]`.

### Answer

```go
ports := make([]int, count)
for i := range count {
	ports[i] = start + i
}
return ports
```

When `count` is `0`, `make` returns an empty slice and the loop never runs.
