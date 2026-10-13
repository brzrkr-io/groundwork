## Task 1: Count the batches

Dividing `len(hosts)` by `size` rounds down: 5 hosts in batches of 2 gives `2`,
one batch short. The README's Learn section shows how to round up instead.

### Closer

Add `size - 1` before dividing: `(len(hosts) + ...) / size`.

### Answer

```go
return (len(hosts) + size - 1) / size
```

Adding `size - 1` pushes any leftover hosts into one more batch, and 0 hosts
still gives `0`.

## Task 2: Pick one batch

Batch `n` starts at index `n*size` and ends `size` hosts later. Either bound can
run past `len(hosts)`, which would panic, so keep both of them in range.

### Closer

Clamp each bound with the built-in `min`, then slice:

```go
start := min(n*size, len(hosts))
end := min(..., len(hosts))
```

### Answer

```go
start := min(n*size, len(hosts))
end := min(start+size, len(hosts))
return hosts[start:end]
```

Once both bounds are at most `len(hosts)`, `hosts[start:end]` never panics, and
a batch past the end becomes `hosts[len(hosts):len(hosts)]`, which is empty.

## Task 3: Plan the whole rollout

You already wrote the two pieces you need: `BatchCount` tells you how many
batches there are, and `Batch` gives you any one of them.

### Closer

Make the outer slice with one element per batch, then fill each element in a
loop:

```go
plan := make([][]string, BatchCount(hosts, size))
for n := range len(plan) {
	plan[n] = ...
}
```

### Answer

```go
plan := make([][]string, BatchCount(hosts, size))
for n := range len(plan) {
	plan[n] = Batch(hosts, size, n)
}
return plan
```

`make` creates exactly as many slots as batches, and each slot is set to its
batch by number.
