## Task 1: Add up the latencies

You need the values in `ms`, not their positions. Range over the slice and add
each value to `total`.

### Closer

Put `_` where the index would go, so you only get the element:
`for _, v := range ms { ... }`.

### Answer

```go
for _, v := range ms {
	total += v
}
return total
```

## Task 2: Find the slowest host

Range over `hosts` with both the index and the host. The latency for that host
is `ms[i]`. Remember the slowest one you've seen so far.

### Closer

Use `for i, host := range hosts`. Replace `slowest` and `worst` when this is
the first host (`i == 0`) or when `ms[i]` is strictly greater (`>`) than
`worst`.

### Answer

```go
for i, host := range hosts {
	if i == 0 || ms[i] > worst {
		slowest, worst = host, ms[i]
	}
}
return slowest, worst
```

`>` (not `>=`) keeps the first host on a tie, and `i == 0` makes sure a host is
picked even when every latency is 0.

## Task 3: List the unhealthy hosts

Range over `healthy`. When a check is `false`, the host at the same index
failed, so add it to `down`.

### Closer

Use `for i, ok := range healthy`. Check `!ok`, then
`down = append(down, ...)` with the host at index `i`.

### Answer

```go
for i, ok := range healthy {
	if !ok {
		down = append(down, hosts[i])
	}
}
return down
```
