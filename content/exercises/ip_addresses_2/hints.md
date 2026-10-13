## Task 1: one method call

`p.String()` prints the prefix as written. Call `Masked()` first, then
`String()` on the result: `p.Masked().String()`.

## Task 2: where does the inner loop start?

With `j := 0`, the inner loop visits every CIDR, including `i` itself and
the ones before it. You only want the CIDRs *after* `i`:

```go
for j := i + 1; j < len(prefixes); j++ {
```

## Task 3: check the family, then shift

`p.Addr()` gives the address part of the prefix, and `Is4()` tells you
whether it is IPv4. A /24 has `32 - 24 = 8` free bits, and `1 << 8` is 256.

```go
if !p.Addr().Is4() {
	return 0, fmt.Errorf("size: CIDR is not IPv4")
}
return 1 << (32 - p.Bits()), nil
```
