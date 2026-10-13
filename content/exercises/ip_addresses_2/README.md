# CIDRs: normalize and audit subnet rules

A firewall audit script reads subnet rules written by many different people.
Fix the code that cleans them up and finds conflicting ranges.

## Learn

### Recap

`netip.ParsePrefix("10.0.0.0/24")` turns a CIDR string into a
`netip.Prefix`, or returns an error if the text is not a valid CIDR. The
number after the slash is how many leading bits are fixed, so `/24` covers
256 IPv4 addresses. `a.Overlaps(b)` reports whether two prefixes share any
address.

### Masking a prefix

People often write a CIDR with "leftover" bits set, like `10.1.2.3/16`.
`ParsePrefix` accepts it, and `String()` prints it back exactly as written.
`Masked()` returns the same range with the free bits set to zero, which is
the standard way to write it:

```go
p, _ := netip.ParsePrefix("10.1.2.3/16")
p.String()          // => "10.1.2.3/16"
p.Masked().String() // => "10.1.0.0/16"
```

### Prefix parts

| Method | Returns | For `10.1.0.0/16` |
|---|---|---|
| `p.Bits()` | the number after the slash, as an `int` | `16` |
| `p.Addr()` | the address part, as a `netip.Addr` | `10.1.0.0` |

### Powers of two with `<<`

`1 << n` shifts the bits of `1` left `n` places, which gives 2 to the power
of `n`: `1 << 0` is 1, `1 << 3` is 8, `1 << 8` is 256. An IPv4 prefix has
`32 - Bits()` free bits, so it holds `1 << (32 - p.Bits())` addresses.

## Your task

The audit script has two bugs and one unfinished function.

### 1. Fix NormalizeCIDR

`NormalizeCIDR(s)` should return the standard form of the CIDR, with the
free bits zeroed. Right now it returns the input unchanged. Keep returning
`""` and the error for invalid input.

```go
NormalizeCIDR("10.1.2.3/16")    // => "10.1.0.0/16", nil
NormalizeCIDR("192.168.7.0/24") // => "192.168.7.0/24", nil
NormalizeCIDR("10.1.2.3")       // => "", error (no slash)
```

### 2. Fix FindOverlaps

`FindOverlaps(cidrs)` returns one line `"A overlaps B"` for each pair of
overlapping CIDRs, with A earlier in the list than B. Each pair appears once,
in list order: all pairs for the first CIDR, then for the second, and so on.
If any CIDR is invalid, it returns `nil` and the error.

Right now it compares each CIDR with itself and reports every pair twice.

```go
FindOverlaps([]string{"10.0.0.0/16", "10.0.1.0/24", "192.168.0.0/16", "10.0.1.128/25"})
// => []string{
//      "10.0.0.0/16 overlaps 10.0.1.0/24",
//      "10.0.0.0/16 overlaps 10.0.1.128/25",
//      "10.0.1.0/24 overlaps 10.0.1.128/25",
//    }, nil

FindOverlaps([]string{"10.0.0.0/16", "10.1.0.0/16"})
// => nil, nil (no overlaps)
```

### 3. Finish Size

`Size(cidr)` returns how many addresses an IPv4 CIDR holds. The parsing is
done. Return an error when the prefix is IPv6 (its sizes don't fit in an
`int`), using `fmt.Errorf` with the message `size: CIDR is not IPv4`.

```go
Size("10.0.0.0/24")   // => 256, nil
Size("10.0.0.0/16")   // => 65536, nil
Size("10.0.0.7/32")   // => 1, nil
Size("2001:db8::/32") // => 0, error "size: CIDR is not IPv4"
```

## Learn more

- [netip.Prefix](https://pkg.go.dev/net/netip#Prefix)
- [Go spec: Arithmetic operators (shifts)](https://go.dev/ref/spec#Arithmetic_operators)
