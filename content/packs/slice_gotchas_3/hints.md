## Task 1: Drain a host

Don't remove anything from `pool` itself. Start a new, empty slice and add
only the hosts you want to keep; a new slice gets its own backing array, so
nothing you do to it can reach `pool`.

### Closer

Declare `var out []string`, loop with `for _, h := range pool`, and `append`
`h` to `out` when it is not the host being drained.

### Answer

```go
var out []string
for _, h := range pool {
	if h != host {
		out = append(out, h)
	}
}
return out
```

`out` starts empty and grows its own array, so `pool` is never written to.

## Task 2: Promote a replica

First handle the case where `host` isn't in the pool: return a copy. Otherwise
the answer is `host` followed by every other host, and `Drain` already gives
you "every other host".

### Closer

`slices.Contains(pool, host)` tells you if `host` is there, and
`slices.Clone(pool)` makes a copy. Start a new slice `[]string{host}` and
append a whole slice to it with `...`. Add `"slices"` to the imports.

### Answer

```go
if !slices.Contains(pool, host) {
	return slices.Clone(pool)
}
return append([]string{host}, Drain(pool, host)...)
```

Both paths build a brand-new slice, so changing the result never changes
`pool`.
