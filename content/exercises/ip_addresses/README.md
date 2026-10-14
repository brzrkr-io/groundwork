# IP addresses: classify, allowlist, overlap

*Project: ops inventory CLI* closed *Packages and project layout*: a
multi-package tool that works on local data. This module adds the network:
addresses, connections and HTTP.

Firewall rules, security groups, VPC subnets and Kubernetes network policies
are all written as IP addresses and ranges. Go's `net/netip` package lets you
check them in code instead of by eye.

**New here:** the `net/netip` package: `netip.Addr`, `netip.Prefix`, CIDR notation.

## Learn

### IP addresses

An *IP address* identifies a machine on a network. There are two kinds:

| Family | Example | Written as |
|---|---|---|
| IPv4 | `192.168.1.10` | four numbers from 0 to 255, separated by dots |
| IPv6 | `2001:db8::1` | groups of hex digits separated by colons |

An IPv4 address is 32 *bits* (ones and zeros) long. An IPv6 address is 128.

### Parsing an address

`netip.ParseAddr` turns a string into a `netip.Addr`. Like `strconv.Atoi`, it
also returns an error when the text is not valid:

```go
addr, err := netip.ParseAddr("172.20.0.9")
if err != nil {
	return err // for example "999.1.1.1" or "localhost"
}
```

> **Why is it written like this?** Returning `(value, error)` and checking
> `err != nil` straight away is a Go *convention*; the compiler doesn't force
> it, but every Go programmer expects it. The names `addr` and `err` are the
> author's choice (`err` is the usual one).

A `netip.Addr` has methods that answer yes/no questions. Each returns a `bool`:

| Method | True when |
|---|---|
| `addr.Is4()` | it is IPv4 |
| `addr.Is6()` | it is IPv6 |
| `addr.IsLoopback()` | it means "this machine": anything in `127.0.0.0/8`, or `::1` |
| `addr.IsPrivate()` | it is in one of the private ranges below |

### Private and public addresses

Private ranges are for internal networks (an office, a cloud VPC). The
internet can't reach them directly. In this exercise, anything that is
neither private nor loopback is *public*.

| Range | Family | Typical use |
|---|---|---|
| `10.0.0.0/8` | IPv4 | big corporate and cloud networks |
| `172.16.0.0/12` | IPv4 | Docker's default networks |
| `192.168.0.0/16` | IPv4 | home and office routers |
| `fc00::/7` | IPv6 | private IPv6 networks |

Loopback addresses are not private, so check loopback *first*.

### CIDR notation

A range like `10.0.0.0/24` uses *CIDR notation*: an address, a slash, and how
many leading bits are fixed. The rest can be anything.

- `/24`: `10.0.0` fixed, covers `10.0.0.0`–`10.0.0.255` (256 addresses).
- `/16`: `10.0` fixed, covers 65,536 addresses.
- `/32`: all bits fixed, exactly one address.

A smaller number means a bigger range. In Go a range is a `netip.Prefix`:

```go
office, err := netip.ParsePrefix("172.16.0.0/12")
if err != nil {
	return err // for example "10.0.0.0/33" or "10.0.0.0" (no slash)
}
laptop, _ := netip.ParseAddr("172.20.1.4")
office.Contains(laptop) // => true
```

`Contains` returns `false` when the families differ: an IPv6 address is never
inside an IPv4 range.

> **Why is it written like this?** `office.Contains(laptop)` is a *method
> call*: the value before the dot is the thing you ask, the value in brackets
> is what you ask about. Go requires that shape. The `_` throws away a value
> you don't need; only do that in examples where you know the input is valid.

### Overlapping ranges

Two prefixes *overlap* when at least one address is in both. `x.Overlaps(y)`
checks that. You can't connect (*peer*) two VPCs whose ranges overlap,
because routers couldn't tell which side an address belongs to.

## Your task

You're building a small network audit tool for your team.

### 1. Classify an address

**Goal:** make `Classify(s)` return `"FAMILY KIND"`, for example `"ipv4 private"`.

- FAMILY is `ipv4` or `ipv6`. The starter already works this out.
- KIND is `loopback` if the address is loopback, otherwise `private` if it is
  private, otherwise `public`.
- Invalid input returns `""` and the parse error. The starter already does this.

```go
Classify("127.0.0.53")           // => "ipv4 loopback", nil
Classify("172.17.0.2")           // => "ipv4 private", nil
Classify("8.8.8.8")              // => "ipv4 public", nil
Classify("::1")                  // => "ipv6 loopback", nil
Classify("fd00::5")              // => "ipv6 private", nil
Classify("2001:4860:4860::8888") // => "ipv6 public", nil
Classify("10.0.0.0/8")           // => "", error (a range, not an address)
```

**Think it through:**
- Which variable holds the parsed address, and which variable holds the word you change?
- Which method from the table answers each question, and in what order should you ask?

### 2. Check an allowlist

**Goal:** make `Allowed(ip, allowlist)` return `true` if `ip` is inside any CIDR in `allowlist`.

- If `ip` is not a valid address, return `false` and the error.
- Go through `allowlist` in order. If a CIDR doesn't parse, return `false` and the error.
- An empty (or `nil`) allowlist allows nothing: `false, nil`.

```go
Allowed("10.0.3.7", []string{"192.168.0.0/16", "10.0.0.0/8"}) // => true, nil
Allowed("8.8.8.8", []string{"10.0.0.0/8"})                    // => false, nil
Allowed("::1", []string{"0.0.0.0/0"})                         // => false, nil
Allowed("10.0.3", []string{"10.0.0.0/8"})                     // => false, error
Allowed("10.0.3.7", []string{"10.0.0.0/99"})                  // => false, error
```

**Think it through:**
- `ip` and each list entry are strings; which parse function fits each one?
- When can you stop looping early, and what do you return if the loop finishes without a match?

### 3. Check two VPC ranges for overlap

**Goal:** make `Overlaps(a, b)` return whether the two CIDRs share any address.

If either one doesn't parse, return `false` and the error.

```go
Overlaps("10.0.0.0/16", "10.0.5.0/24") // => true, nil
Overlaps("10.0.5.0/24", "10.0.0.0/16") // => true, nil
Overlaps("10.0.0.0/16", "10.1.0.0/16") // => false, nil
Overlaps("10.0.0.0/16", "banana")      // => false, error
```

**Think it through:**
- You need two parsed prefixes; check the error after each parse.
- Which prefix method from the Learn section answers "do these share an address"?

## Learn more

- [net/netip package](https://pkg.go.dev/net/netip)
- [Go blog: netip](https://go.dev/blog/netip)
- [Amazon VPC: IP addressing](https://docs.aws.amazon.com/vpc/latest/userguide/vpc-ip-addressing.html)
