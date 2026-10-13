## Task 1: round up

`len(hosts) / size` rounds down, so 5 hosts in batches of 2 gives `2`, one
batch short. Use `(len(hosts) + size - 1) / size` instead.

## Task 2: work out start and end

Batch `n` starts at `n*size` and ends at `n*size + size`. Either bound can run
past `len(hosts)`, so clamp both with `min(..., len(hosts))`. Once `start` and
`end` are clamped, `hosts[start:end]` never panics.

## Task 3: reuse your functions

`make([][]string, BatchCount(hosts, size))` gives you one empty element per
batch. Loop over the indexes and set each element with `Batch`.

## Almost the answer

```go
func Batch(hosts []string, size, n int) []string {
	start := min(n*size, len(hosts))
	end := min(start+size, len(hosts))
	return hosts[start:end]
}
```

```go
plan := make([][]string, BatchCount(hosts, size))
for n := range len(plan) {
	plan[n] = Batch(hosts, size, n)
}
return plan
```
