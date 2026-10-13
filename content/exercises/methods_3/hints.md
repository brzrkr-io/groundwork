## Loop over the receiver's slice

Every method starts the same way: `for _, h := range inv.Hosts { ... }`.
Inside the loop, `h` is one host, so you can check `h.Healthy`, `h.Region` and
`h.Name`.

## Task 1 and 2: count and collect

For `HealthyCount`, start a counter at 0 and add 1 for every healthy host.
For `InRegion`, start with `var hosts []Host` and `append` each matching
host. Return the counter or the slice after the loop.

## Task 3: return early

Compare names with `strings.EqualFold(h.Name, name)` and return as soon as one
matches. Only return `Host{}, false` after the loop has checked every host.

## Almost the answer

```go
for _, h := range inv.Hosts {
	if strings.EqualFold(h.Name, name) {
		return h, true
	}
}
return Host{}, false
```
