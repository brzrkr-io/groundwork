## Task 1: Fix the blank addresses

The bug is in the `make` line. Ask how many elements `out` holds *before* the
loop runs, and remember that `append` adds after whatever is already there.

### Closer

`make([]string, len(hosts))` already holds `len(hosts)` empty strings. Start
with length 0 and pass the capacity as the third argument:
`make([]string, 0, ...)`.

### Answer

```go
out := make([]string, 0, len(hosts))
```

Length 0 means `append` starts at the front; the capacity just saves regrowing.

## Task 2: Fix the empty backup

The bug is again in the `make` line. `copy` only fills slots that already
exist, so look at the length of `backup`.

### Closer

`copy` copies the smaller of `len(dst)` and `len(src)`. A destination of length
0 gets nothing, so give `backup` the same *length* as `hosts`.

### Answer

```go
backup := make([]string, len(hosts))
```

Now `backup` has a slot for every host, so `copy` fills them all.
