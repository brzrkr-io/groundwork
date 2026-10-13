## Build, don't edit

Don't remove anything from `pool` itself. Start a new, empty slice and
`append` only the hosts you want. That new slice gets its own backing array,
so nothing you do to it can reach `pool`.

## Task 1: a filtering loop

```go
var out []string
for _, h := range pool {
	if h != host {
		out = append(out, h)
	}
}
return out
```

## Task 2: front first, then the rest

If `slices.Contains(pool, host)` is false, return `slices.Clone(pool)`.
Otherwise start a new slice holding just `host`, and append everything
`Drain(pool, host)` returns:

```go
return append([]string{host}, Drain(pool, host)...)
```

Add `"slices"` to the imports.
