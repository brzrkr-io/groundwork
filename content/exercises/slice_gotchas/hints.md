## Who else points at this array?

Each starter returns a slice that still points into its input's backing
array. Draw the boxes from the README: which slices share the array, and where
would `append` write next?

## Task 1: clone it

`slices.Clone` makes a new backing array:

```go
return slices.Clone(hosts[:n])
```

Add `"slices"` to the imports.

## Task 2: take away the spare capacity

`base[:len(base):len(base)]` is all of `base` with `cap == len`. Appending to
that has no room, so `append` copies into a new array and leaves `base`'s
array alone:

```go
return append(base[:len(base):len(base)], extra)
```

## Task 3: make + copy

```go
out := make([]string, n) // len n, cap n, new array
copy(out, lines[len(lines)-n:])
return out
```
