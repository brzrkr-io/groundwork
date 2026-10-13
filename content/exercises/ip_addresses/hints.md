## Task 1: ask the address

`addr` already holds the parsed address. Its methods `IsLoopback()` and
`IsPrivate()` answer the questions for you. Check loopback first, because it
wins over private.

```go
if addr.IsLoopback() {
	kind = "loopback"
} else if addr.IsPrivate() {
	kind = "private"
}
```

## Task 2: parse, then loop

Start like `Classify`: `netip.ParseAddr(ip)`, and return the error if there
is one. Then loop over `allowlist`, turning each string into a prefix with
`netip.ParsePrefix(cidr)`. Return `true, nil` as soon as
`prefix.Contains(addr)` is true. After the loop, return `false, nil`.

## Task 3: two prefixes

Parse `a` and `b` with `netip.ParsePrefix`, checking the error after each
one. Then:

```go
return pa.Overlaps(pb), nil
```

## Almost the answer for task 2

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
