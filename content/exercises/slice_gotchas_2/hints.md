## Draw the boxes

In both functions, two slices end up pointing at the same backing array, and
one of them still has spare capacity. Find the slice that `append` writes
through, and ask what its `cap` is.

## Task 1: limit the canary's capacity

`hosts[:n]` has `cap(hosts)` room, so appending to it writes over `hosts[n]`,
which is `rest[0]`. A full slice expression `hosts[:n:n]` gives the canary
`cap == n`.

## Task 2: a prefix with no room

`args` was made with `cap 8`, so `append(args, ...)` writes into box 5 of the
same array, both times. Take a view of `args` with no spare capacity and
append to that instead:

```go
prefix := args[:len(args):len(args)]
canary := append(prefix, "deployment/"+deployment+"-canary")
main := append(prefix, "deployment/"+deployment)
```
