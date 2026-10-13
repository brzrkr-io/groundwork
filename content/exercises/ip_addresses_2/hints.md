## Task 1: Fix NormalizeCIDR

`p.String()` prints the prefix exactly as written. Look in the Learn section
for the method that zeroes the free bits.

### Closer

Call `Masked()` first, then `String()` on what it returns.

### Answer

```go
return p.Masked().String(), nil
```

## Task 2: Fix FindOverlaps

Look at where the inner loop starts. With `j := 0` it visits `i` itself and
every CIDR before it, so it compares a CIDR with itself and sees each pair
twice.

### Closer

You only want the CIDRs *after* `i`, so start `j` one past it.

### Answer

```go
for j := i + 1; j < len(prefixes); j++ {
```

## Task 3: Finish Size

First check that the prefix is IPv4, then work out how many free bits it has.
A /24 has `32 - 24 = 8` free bits, and `1 << 8` is 256.

### Closer

`p.Addr().Is4()` tells you the family, and `p.Bits()` is the number after the
slash: `1 << (32 - p.Bits())`.

### Answer

```go
if !p.Addr().Is4() {
	return 0, fmt.Errorf("size: CIDR is not IPv4")
}
return 1 << (32 - p.Bits()), nil
```
