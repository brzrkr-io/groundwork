## Task 1: Count healthy hosts

Loop over `inv.Hosts` and keep a running count. Inside the loop, each host's
`Healthy` field tells you whether to count it.

### Closer

Start with `count := 0`, loop with `for _, h := range inv.Hosts { ... }`, and
inside use `if h.Healthy { count++ }`. Return `count` after the loop.

### Answer

```go
count := 0
for _, h := range inv.Hosts {
	if h.Healthy {
		count++
	}
}
return count
```

## Task 2: Filter by region

Same loop as Task 1, but instead of counting you collect the matching hosts
into a new slice.

### Closer

Start with `var hosts []Host` (a nil slice is fine when nothing matches), and
inside the loop use `hosts = append(hosts, h)` when `h.Region == region`.

### Answer

```go
var hosts []Host
for _, h := range inv.Hosts {
	if h.Region == region {
		hosts = append(hosts, h)
	}
}
return hosts
```

## Task 3: Find a host by name

Loop over the hosts and stop as soon as one has a matching name. Only after
the loop has checked every host do you know there's no match.

### Closer

Compare with `strings.EqualFold(h.Name, name)` (add `"strings"` to the
imports) and `return h, true` inside the loop. Put `return Host{}, false`
after the loop.

### Answer

```go
for _, h := range inv.Hosts {
	if strings.EqualFold(h.Name, name) {
		return h, true
	}
}
return Host{}, false
```
