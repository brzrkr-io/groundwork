## Task 1: Split the canary wave

`canary` and `rest` share one backing array, and `canary` still has spare
capacity. When you `append` to `canary`, which box of the array does it write
into?

### Closer

`hosts[:n]` has `cap(hosts)` room, so its next `append` writes over `hosts[n]`,
which is `rest[0]`. A full slice expression `s[low:high:max]` limits the
capacity: make `max` equal to `high`.

### Answer

```go
return hosts[:n:n], hosts[n:]
```

With `cap == n`, appending to `canary` must copy into a new array, so `rest`
and `hosts` stay untouched.

## Task 2: Build both restart commands

`args` was made with `cap 8` but only holds 5 items. Both `append` calls write
into the same spare box (index 5), so the second one overwrites the first.

### Closer

Take a view of `args` that has no spare capacity, using a full slice
expression `args[:len(args):len(args)]`, and append to that view instead of
to `args`.

### Answer

```go
prefix := args[:len(args):len(args)]
canary := append(prefix, "deployment/"+deployment+"-canary")
main := append(prefix, "deployment/"+deployment)
```

`prefix` is full, so each `append` copies into its own new array.
