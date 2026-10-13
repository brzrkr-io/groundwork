## Task 1: Copy the canary hosts

`hosts[:n]` is a new header pointing into the same array, so changing it
changes `hosts`. You need a new backing array with the same elements.

### Closer

`slices.Clone(s)` makes a copy with its own array. Clone the part you want:
`slices.Clone(hosts[:...])`. Add `"slices"` to the imports.

### Answer

```go
return slices.Clone(hosts[:n])
```

The clone has its own array, so writing to it leaves `hosts` alone.

## Task 2: Add a step without overwriting

`base` has spare capacity, so `append` writes into the next box of `base`'s
array, and the second call overwrites the first. Take away the spare capacity
before you append.

### Closer

The full slice expression `s[low:high:max]` limits `cap` to `max - low`.
`base[:len(base):len(base)]` is all of `base` with `cap == len`, so `append`
has no room and must copy into a new array.

### Answer

```go
return append(base[:len(base):len(base)], extra)
```

With no spare capacity, `append` allocates a new array instead of writing into
`base`'s.

## Task 3: Keep only the tail of a log

`lines[len(lines)-n:]` still points into the big log array. Make a brand-new
slice of exactly `n` elements and fill it from the tail.

### Closer

`make([]string, n)` gives a new array with `len` and `cap` both `n`.
`copy(dst, src)` copies elements from `src` into `dst`; the source here is
`lines[len(lines)-n:]`.

### Answer

```go
out := make([]string, n)
copy(out, lines[len(lines)-n:])
return out
```

`out` has its own array of exactly `n` boxes, so the big log can be freed.
