## Task 1: Classify the lookup error

Handle `nil` first. Then dig out the `*net.DNSError`: if there isn't one, the
answer is `Failed`. If there is, check its fields in the order the task lists.

### Closer

`dnsErr, ok := errors.AsType[*net.DNSError](err)` finds it. A `switch` with no
value after the keyword checks `case !ok:`, `case dnsErr.IsNotFound:` and so
on, top to bottom, as in *Switch: size, method, latency*.

### Answer

```go
if err == nil {
	return OK
}
dnsErr, ok := errors.AsType[*net.DNSError](err)
switch {
case !ok:
	return Failed
case dnsErr.IsNotFound:
	return NotFound
case dnsErr.IsTimeout:
	return Timeout
case dnsErr.IsTemporary:
	return ServFail
default:
	return Failed
}
```

## Task 2: Split A and AAAA answers

Loop over `addrs` and parse each one. Skip it when the parse fails; otherwise
the parsed address tells you which slice the original string goes into.

### Closer

`addr, err := netip.ParseAddr(s)`, then `continue` on an error. Append `s`
(the string, not `addr`) with `v4 = append(v4, s)` when `addr.Is4()` is true.

### Answer

```go
for _, s := range addrs {
	addr, err := netip.ParseAddr(s)
	if err != nil {
		continue
	}
	if addr.Is4() {
		v4 = append(v4, s)
	} else {
		v6 = append(v6, s)
	}
}
return v4, v6
```

## Task 3: Look up with a timeout

Three steps: make the context (and defer its `cancel`), call `r.LookupHost`,
then build the `Result` from the two helpers you already wrote.

### Closer

`ctx, cancel := context.WithTimeout(context.Background(), timeout)` followed
by `defer cancel()`. Then `addrs, err := r.LookupHost(ctx, host)`.
`SplitFamilies(nil)` already returns two `nil` slices, so no special case.

### Answer

```go
ctx, cancel := context.WithTimeout(context.Background(), timeout)
defer cancel()
addrs, err := r.LookupHost(ctx, host)
v4, v6 := SplitFamilies(addrs)
return Result{Host: host, IPv4: v4, IPv6: v6, Status: Classify(err)}
```
