# CIDRs: normalize and audit subnet rules

A firewall audit script reads subnet rules written by many different people.
Fix the code that cleans them up and finds conflicting ranges.

**New here:** `Prefix.Masked`, `Prefix.Bits`, `Prefix.Addr`, the `<<` shift operator.

## Learn

### Recap

`netip.ParsePrefix` turns a CIDR string like `"10.0.0.0/24"` into a
`netip.Prefix`, or returns an error. The number after the slash is how many
leading bits are fixed. `x.Overlaps(y)` reports whether two prefixes share
an address.

### Masking a prefix

People often write a CIDR with leftover bits set, like `172.16.99.1/12`.
`ParsePrefix` accepts it, and `String()` prints it back exactly as written.
`Masked()` returns a *new* prefix with the free bits set to zero, the
standard way to write it:

```go
rule, _ := netip.ParsePrefix("172.16.99.1/12")
fmt.Println(rule)  // 172.16.99.1/12
clean := rule.Masked()
fmt.Println(clean) // 172.16.0.0/12
```

> **Why is it written like this?** `Masked()` returns a `netip.Prefix`, not a
> string, so you can call another method on its result. Go allows chaining
> calls with dots like `a.B().C()`; whether you chain or use a variable such
> as `clean` is the author's choice.

### Prefix parts

| Method | Returns | For `10.1.0.0/16` |
|---|---|---|
| `p.Bits()` | the number after the slash, as an `int` | `16` |
| `p.Addr()` | the address part, as a `netip.Addr` | `10.1.0.0` |

`Addr()` gives you a `netip.Addr`, so its methods (`Is4()`, `Is6()`) work on
the result.

### Powers of two with `<<`

`1 << n` shifts the bits of `1` left `n` places. That gives 2 to the power
of `n`: `1 << 0` is 1, `1 << 3` is 8, `1 << 10` is 1024.

An IPv4 address has 32 bits. The bits that are *not* fixed are free, and each
free bit doubles the number of addresses.

> **Why is it written like this?** Go has no `**` power operator; `<<` is the
> usual way to get a power of two. Put brackets around a subtraction on the
> right of `<<`: Go's `<<` binds tighter than `-`, so `1 << 32 - 8` means
> `(1 << 32) - 8`.

## Your task

The audit script has two bugs and one unfinished function.

### 1. Fix NormalizeCIDR

**Goal:** make `NormalizeCIDR(s)` return the CIDR with its free bits zeroed.

Right now it returns the input unchanged. Keep returning `""` and the error
for invalid input. IPv6 works the same way.

```go
NormalizeCIDR("10.1.2.3/16")    // => "10.1.0.0/16", nil
NormalizeCIDR("192.168.7.0/24") // => "192.168.7.0/24", nil
NormalizeCIDR("2001:db8::1/32") // => "2001:db8::/32", nil
NormalizeCIDR("10.1.2.3")       // => "", error (no slash)
```

**Think it through:**
- The bug is on the `return` line marked `TODO(1)`. What does it print right now?
- Which method from the Learn section zeroes the free bits?

### 2. Fix FindOverlaps

**Goal:** make `FindOverlaps(cidrs)` report each overlapping pair exactly once, as `"A overlaps B"`.

- A comes earlier in the list than B.
- Order: all pairs for the first CIDR, then for the second, and so on.
- Print the CIDRs as they were written (not normalized).
- No overlaps returns `nil, nil`. Any invalid CIDR returns `nil` and the error (already done).

Right now it compares each CIDR with itself and reports every pair twice.

```go
FindOverlaps([]string{"10.0.0.0/16", "10.0.1.0/24", "192.168.0.0/16", "10.0.1.128/25"})
// => []string{
//      "10.0.0.0/16 overlaps 10.0.1.0/24",
//      "10.0.0.0/16 overlaps 10.0.1.128/25",
//      "10.0.1.0/24 overlaps 10.0.1.128/25",
//    }, nil

FindOverlaps([]string{"10.0.0.0/16", "10.1.0.0/16"}) // => nil, nil
```

**Think it through:**
- Look at where the inner loop's `j` starts. Which positions does it visit?
- Which positions should `j` visit, compared with `i`, so each pair appears once?

### 3. Finish Size

**Goal:** make `Size(cidr)` return how many addresses an IPv4 CIDR holds.

- Parsing is done for you.
- If the prefix is IPv6, return `0` and an error made with `fmt.Errorf`, message exactly `size: CIDR is not IPv4`.
- Right now it returns the number of bits instead of the number of addresses.

```go
Size("10.0.0.0/24")   // => 256, nil
Size("10.0.0.0/28")   // => 16, nil
Size("10.0.0.7/32")   // => 1, nil
Size("2001:db8::/32") // => 0, error "size: CIDR is not IPv4"
```

**Think it through:**
- How do you get from the prefix to an address you can ask "are you IPv4?"
- A `/24` has `32 - 24 = 8` free bits. How does `<<` turn free bits into a count?

## Learn more

- [netip.Prefix](https://pkg.go.dev/net/netip#Prefix)
- [Go spec: Arithmetic operators (shifts)](https://go.dev/ref/spec#Arithmetic_operators)
- [Go spec: Operator precedence](https://go.dev/ref/spec#Operator_precedence)
