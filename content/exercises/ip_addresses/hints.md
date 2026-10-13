## Task 1: Classify an address

`addr` already holds the parsed address, and its methods answer the questions
for you. Check loopback first, because it wins over private.

### Closer

`addr.IsLoopback()` and `addr.IsPrivate()` return a `bool`. Use an
`if` / `else if` that changes `kind`, which starts as `"public"`.

### Answer

```go
kind := "public"
if addr.IsLoopback() {
	kind = "loopback"
} else if addr.IsPrivate() {
	kind = "private"
}
```

## Task 2: Check an allowlist

Parse `ip` first, as `Classify` does. Then go through the allowlist one CIDR
at a time and stop as soon as one contains the address.

### Closer

`netip.ParsePrefix(cidr)` turns each string into a prefix, and
`prefix.Contains(addr)` checks the address. Return `true, nil` inside the loop
and `false, nil` after it.

### Answer

```go
addr, err := netip.ParseAddr(ip)
if err != nil {
	return false, err
}
for _, cidr := range allowlist {
	prefix, err := netip.ParsePrefix(cidr)
	if err != nil {
		return false, err
	}
	if prefix.Contains(addr) {
		return true, nil
	}
}
return false, nil
```

## Task 3: Check two VPC ranges for overlap

Parse both CIDRs, checking the error after each one. Then one method call
answers the question.

### Closer

`netip.ParsePrefix(a)` and `netip.ParsePrefix(b)`, then `pa.Overlaps(pb)`.

### Answer

```go
pa, err := netip.ParsePrefix(a)
if err != nil {
	return false, err
}
pb, err := netip.ParsePrefix(b)
if err != nil {
	return false, err
}
return pa.Overlaps(pb), nil
```
