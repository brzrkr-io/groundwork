# IP addresses: classify, allowlist, overlap

Firewall rules, security groups, VPC subnets and Kubernetes network policies
are all written as IP addresses and address ranges. Go's `net/netip` package
lets you check them in code instead of by eye.

## Learn

### IP addresses

An *IP address* identifies a machine on a network. There are two kinds:

| Family | Example | Written as |
|---|---|---|
| IPv4 | `192.168.1.10` | four numbers from 0 to 255, separated by dots |
| IPv6 | `2001:db8::1` | groups of hex digits separated by colons |

Every IPv4 address is 32 *bits* (ones and zeros) long; every IPv6 address is
128 bits long.

### Parsing an address

`netip.ParseAddr` turns a string into a `netip.Addr`. Like `strconv.Atoi`, it
returns an error when the text is not valid:

```go
import "net/netip"

addr, err := netip.ParseAddr("10.0.0.5")
if err != nil {
	return err // for example "999.1.1.1" or "localhost"
}
```

A `netip.Addr` has methods that answer yes/no questions:

| Method | True when |
|---|---|
| `addr.Is4()` | it is an IPv4 address |
| `addr.Is6()` | it is an IPv6 address |
| `addr.IsLoopback()` | it means "this machine": `127.0.0.1` (all of `127.0.0.0/8`) or `::1` |
| `addr.IsPrivate()` | it is in one of the private ranges below |

### Private and public addresses

Some ranges are reserved for private networks such as your office or a cloud
VPC. They are never reachable directly from the internet. Everything that is
neither private nor loopback, we'll call *public*.

| Range | Family | Typical use |
|---|---|---|
| `10.0.0.0/8` | IPv4 | big corporate and cloud networks |
| `172.16.0.0/12` | IPv4 | Docker's default networks |
| `192.168.0.0/16` | IPv4 | home and office routers |
| `fc00::/7` | IPv6 | private IPv6 networks |

### CIDR notation

A range like `10.0.0.0/24` uses *CIDR notation*: an address, a
slash, and a number of bits. The number says how many of the leading bits are
fixed. The rest can be anything.

- `10.0.0.0/24`: the first 24 bits (`10.0.0`) are fixed, the last 8 bits are
  free, so it covers `10.0.0.0` to `10.0.0.255`: 256 addresses.
- `10.0.0.0/16`: only `10.0` is fixed, covering `10.0.0.0` to
  `10.0.255.255`: 65,536 addresses.
- `10.0.0.7/32`: all 32 bits fixed, exactly one address.

A smaller number after the slash means a bigger range. In Go a CIDR range is
a `netip.Prefix`:

```go
prefix, err := netip.ParsePrefix("10.0.0.0/24")
if err != nil {
	return err // for example "10.0.0.0/33" or "10.0.0.0"
}

addr, _ := netip.ParseAddr("10.0.0.42")
prefix.Contains(addr) // => true
```

`Contains` returns `false` when the address and the prefix are different
families (an IPv6 address is never inside an IPv4 range).

Two prefixes *overlap* when at least one address is in both.
`a.Overlaps(b)` checks that:

```go
a, _ := netip.ParsePrefix("10.0.0.0/16")
b, _ := netip.ParsePrefix("10.0.5.0/24")
a.Overlaps(b) // => true, 10.0.5.0/24 sits inside 10.0.0.0/16
```

Overlaps matter in the cloud: you cannot connect (*peer*) two VPCs whose
address ranges overlap, because the routers could not tell which side an
address belongs to.

## Your task

You're building a small network audit tool for your team. It checks
addresses, allowlists and VPC ranges.

### 1. Classify an address

`Classify(s)` parses `s` and returns `"FAMILY KIND"`:

- FAMILY is `ipv4` or `ipv6`. The starter already works this out.
- KIND is `loopback` if the address is loopback, otherwise `private` if it
  is private, otherwise `public`.

If `s` is not a valid address, return `""` and the error from
`netip.ParseAddr`. The starter already does that too.

```go
Classify("127.0.0.1")    // => "ipv4 loopback", nil
Classify("10.20.30.40")  // => "ipv4 private", nil
Classify("8.8.8.8")      // => "ipv4 public", nil
Classify("::1")          // => "ipv6 loopback", nil
Classify("fd00::5")      // => "ipv6 private", nil
Classify("not-an-ip")    // => "", error
```

### 2. Check an allowlist

`Allowed(ip, allowlist)` reports whether `ip` is inside any CIDR in
`allowlist`.

- If `ip` is not a valid address, return `false` and the error.
- Go through the allowlist in order. If you reach a CIDR that does not parse,
  return `false` and the error.
- An empty allowlist allows nothing.

```go
Allowed("10.0.3.7", []string{"192.168.0.0/16", "10.0.0.0/8"}) // => true, nil
Allowed("8.8.8.8", []string{"10.0.0.0/8"})                    // => false, nil
Allowed("10.0.3.7", []string{"10.0.0.0/99"})                  // => false, error
```

### 3. Check two VPC ranges for overlap

`Overlaps(a, b)` parses both CIDRs and reports whether they overlap. If
either one does not parse, return `false` and the error.

```go
Overlaps("10.0.0.0/16", "10.0.5.0/24")  // => true, nil
Overlaps("10.0.0.0/16", "10.1.0.0/16")  // => false, nil
Overlaps("10.0.0.0/16", "banana")       // => false, error
```

## Learn more

- [net/netip package](https://pkg.go.dev/net/netip)
- [Go blog: netip](https://go.dev/blog/netip)
- [Amazon VPC: IP addressing](https://docs.aws.amazon.com/vpc/latest/userguide/vpc-ip-addressing.html)
