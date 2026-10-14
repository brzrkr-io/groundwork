# VPC planning: subnets per availability zone

Every AWS network starts with a plan: which address range the VPC gets, and
how it is cut into subnets across availability zones. Get it wrong and you
find out months later, when two subnets overlap or a zone runs out of
addresses. Here you write the planner and the checker.

**New here:** VPCs, subnets, public and private subnets, `Addr.As4`, `netip.AddrFrom4`, `netip.PrefixFrom`, `netip.MustParseAddr`/`MustParsePrefix`.

## Learn

### VPCs and subnets

A *VPC* (Virtual Private Cloud) is your own private network inside one AWS
region. You give it an address range in CIDR notation, usually a private
`/16` such as `10.0.0.0/16` (65,536 addresses).

A *subnet* is a slice of the VPC's range that lives in **exactly one**
*availability zone* (AZ): one or more data centres in the region, named
like `eu-west-1a`. Servers (EC2 instances) are always started inside a
subnet. If a zone has a power cut, everything in its subnets goes down, so
you give every zone the same set of subnets and spread copies of each
service across them.

```text
VPC 10.0.0.0/16 (region eu-west-1)
├── eu-west-1a: public  10.0.0.0/20    private 10.0.48.0/20
├── eu-west-1b: public  10.0.16.0/20   private 10.0.64.0/20
├── eu-west-1c: public  10.0.32.0/20   private 10.0.80.0/20
└── 10.0.96.0 – 10.0.255.255: free for later
```

### Public and private subnets

What makes a subnet public is its *route table*: the list of "traffic for
this range goes there" rules attached to the subnet.

| Kind | Route for `0.0.0.0/0` (everything else) | Typical contents |
|---|---|---|
| public | an *internet gateway*: two-way internet access | load balancers, bastion hosts |
| private | a *NAT gateway*, or no route at all | app servers, databases |

A NAT gateway sits in a public subnet and lets private servers make
outgoing connections (to download updates) without the internet being able
to connect to them.

### Size rules and reserved addresses

AWS subnets are between `/16` and `/28` (16 addresses). In every subnet AWS
keeps five addresses for itself. For `10.0.0.0/24`:

| Address | Kept for |
|---|---|
| `10.0.0.0` | network address |
| `10.0.0.1` | the VPC router |
| `10.0.0.2` | the DNS server |
| `10.0.0.3` | future use |
| `10.0.0.255` | broadcast (the last address) |

So a `/24` has 256 − 5 = 251 usable addresses. A subnet of length `/n` has
`1 << (32 - n)` addresses in total, as in *CIDRs: normalize and audit subnet
rules*. The starter's `Reserved` constant holds the 5.

### Addresses as numbers

To find "the third `/20` in a `/16`", you need arithmetic on addresses. An
IPv4 address is really a 32-bit number. `addr.As4()` gives its four bytes,
and `netip.AddrFrom4` builds an address from four bytes:

```go
b := netip.MustParseAddr("192.0.2.10").As4() // [192 0 2 10]
b[3] = 99
netip.AddrFrom4(b) // 192.0.2.99
```

`netip.MustParseAddr` and `netip.MustParsePrefix` work like
`regexp.MustCompile`: they return the value without an error and panic on
bad input. Use them only for fixed values, in examples and tests.

The starter goes one step further with two helpers: `toUint32` glues the
four bytes into one `uint32`, and `fromUint32` cuts it back up. With them,
"200 addresses later" is plain addition:

```go
n := toUint32(netip.MustParseAddr("192.0.2.10"))
fromUint32(n + 200) // 192.0.2.210
```

> **Why is it written like this?** `As4` returns a `[4]byte`: an *array*, a
> list whose length is part of its type. Go requires exactly four bytes, and
> `[4]byte{...}` writes one. `>>` shifts bits right (the opposite of `<<`),
> `byte(...)` keeps only the lowest 8 bits, and `|` puts bits together. You
> use the helpers; you don't have to write them.

### Building a prefix

`netip.PrefixFrom(addr, bits)` makes a prefix from an address and a length:

```go
netip.PrefixFrom(netip.MustParseAddr("192.0.2.64"), 26) // 192.0.2.64/26
```

### Blocks inside a prefix

Cutting a prefix into equal blocks of a longer length works like this
example with `192.168.0.0/24` cut into `/26` blocks:

| Question | Rule | For `/24` into `/26` |
|---|---|---|
| How many blocks? | `1 << (new - old)` | `1 << 2` = 4 |
| How big is one block? | `1 << (32 - new)` | `1 << 6` = 64 addresses |
| Where does block `i` start? | first address + `i` × size | block 3: `192.168.0.192` |

### Try it (optional)

These commands only read. Needs an AWS account and credentials; never
needed for the tests.

```bash
aws ec2 describe-availability-zones --region eu-west-1    # the zone names you can use
aws ec2 describe-vpcs                                     # your VPCs and their CIDRs
aws ec2 describe-subnets --filters Name=vpc-id,Values=vpc-0123456789abcdef0  # one VPC's subnets, zones, free addresses
aws ec2 describe-route-tables --filters Name=vpc-id,Values=vpc-0123456789abcdef0  # which subnets route to an internet gateway
```

Replace `vpc-0123456789abcdef0` with a VPC ID from `describe-vpcs`.

## Your task

Your team is building a new VPC and wants the subnet plan as code, plus a
checker for hand-written plans.

### 1. Carve out one subnet

**Goal:** `SubnetCIDR(vpc, bits, index)` returns block number `index` (counting from 0) of length `/bits` inside `vpc`.

Return an error when:

- `vpc` is not IPv4 (the starter already does this: `As4` would panic on
  an IPv6 address);
- `bits` is shorter than the VPC (`bits < vpc.Bits()`) or longer than `/28`;
- `index` is below 0 or not below the number of blocks.

The starter already starts from `vpc.Masked()`, so leftover bits such as
`10.0.7.9/16` count as `10.0.0.0/16`. Any error text is fine; `fmt` is
already imported for `fmt.Errorf`.

```go
v := netip.MustParsePrefix("10.0.0.0/16")
SubnetCIDR(v, 20, 0)  // => 10.0.0.0/20, nil
SubnetCIDR(v, 20, 2)  // => 10.0.32.0/20, nil
SubnetCIDR(v, 24, 255) // => 10.0.255.0/24, nil
SubnetCIDR(v, 16, 0)  // => 10.0.0.0/16, nil (the whole VPC)
SubnetCIDR(v, 20, 16) // => error: a /16 holds 16 blocks of /20 (0 to 15)
SubnetCIDR(v, 29, 0)  // => error: smaller than /28
SubnetCIDR(netip.MustParsePrefix("2001:db8::/56"), 64, 0) // => error: IPv6
```

**Think it through:**
- Which two rules from *Blocks inside a prefix* give the number of blocks and the size of one block?
- `start` is a `uint32`. What do you convert `index` to before adding?

### 2. Plan subnets for every zone

**Goal:** `Plan(vpc, zones, bits)` returns one public and one private `Subnet` per zone, each `/bits` long.

- Parse `vpc` with `netip.ParsePrefix`; return `nil` and the error if it fails.
- First the public subnets, one per zone in `zones` order, at block
  indexes `0, 1, 2, …`. Then the private ones in the same zone order, at
  the next indexes: `len(zones)`, `len(zones)+1`, ….
- `Name` is `public-<zone>` or `private-<zone>`. `Zone` is the zone.
  `Public` is `true` for the public ones. `Usable` is the block's address
  count minus `Reserved`.
- If `SubnetCIDR` returns an error (subnets too big to fit, or too small),
  return `nil` and that error.
- No zones: no subnets (`nil` is fine) and no error.

```go
Plan("10.0.0.0/16", []string{"eu-west-1a", "eu-west-1b", "eu-west-1c"}, 20)
// => [
//   {public-eu-west-1a  eu-west-1a 10.0.0.0/20  true  4091}
//   {public-eu-west-1b  eu-west-1b 10.0.16.0/20 true  4091}
//   {public-eu-west-1c  eu-west-1c 10.0.32.0/20 true  4091}
//   {private-eu-west-1a eu-west-1a 10.0.48.0/20 false 4091}
//   {private-eu-west-1b eu-west-1b 10.0.64.0/20 false 4091}
//   {private-eu-west-1c eu-west-1c 10.0.80.0/20 false 4091}
// ], nil
Plan("10.0.0.0/16", []string{"eu-west-1a", "eu-west-1b", "eu-west-1c"}, 17)
// => nil, error (six /17s don't fit in a /16)
```

**Think it through:**
- Which index does the private subnet for `zones[i]` get?
- The public and private loops differ only in the name prefix, the index offset and `Public`. One loop inside another over `"public"` and `"private"` saves writing it twice.

### 3. Check a hand-written plan

**Goal:** `Check(vpc, subnets)` returns one message per problem; a good plan gives an empty result (`nil` is fine).

If `vpc` doesn't parse, return just `["invalid VPC CIDR: <vpc>"]`.
Otherwise go through `subnets` in order. For each subnet, add these
messages in this order:

| Problem | Message |
|---|---|
| its length is over `/28` | `<name>: <cidr> is smaller than /28` |
| it is not inside the VPC | `<name>: <cidr> is not inside <vpc>` |
| it overlaps a subnet that comes **later** in the list (one message per later subnet) | `<name> (<cidr>) overlaps <other name> (<other cidr>)` |

A subnet is *inside* the VPC when the VPC contains the subnet's first
address **and** the subnet is at least as long as the VPC: `10.0.0.0/15`
is not inside `10.0.0.0/16`, because it is twice as big. Only looking at
later subnets reports each overlapping pair once.

```go
Check("10.0.0.0/16", []Subnet{
	{Name: "web-a", CIDR: netip.MustParsePrefix("10.0.0.0/24")},
	{Name: "web-b", CIDR: netip.MustParsePrefix("10.0.1.0/24")},
	{Name: "db-a", CIDR: netip.MustParsePrefix("10.0.0.128/25")},
})
// => ["web-a (10.0.0.0/24) overlaps db-a (10.0.0.128/25)"]

Check("172.31.0.0/16", []Subnet{
	{Name: "odd", CIDR: netip.MustParsePrefix("192.168.1.0/30")},
	{Name: "a", CIDR: netip.MustParsePrefix("172.31.0.0/24")},
	{Name: "b", CIDR: netip.MustParsePrefix("172.31.0.0/24")},
})
// => ["odd: 192.168.1.0/30 is smaller than /28"
//     "odd: 192.168.1.0/30 is not inside 172.31.0.0/16"
//     "a (172.31.0.0/24) overlaps b (172.31.0.0/24)"]

Check("vpc-0abc", nil) // => ["invalid VPC CIDR: vpc-0abc"]
```

A `netip.Prefix` prints as `10.0.0.0/24` with `%s`. Use the `vpc` string
as given in the "not inside" message.

**Think it through:**
- Which `netip.Prefix` methods give you the length, the first address, and "do these share an address"?
- How does `subnets[i+1:]` help you look only at the later subnets?

## Learn more

- [How Amazon VPC works](https://docs.aws.amazon.com/vpc/latest/userguide/how-it-works.html)
- [Subnet CIDR blocks (sizes and reserved addresses)](https://docs.aws.amazon.com/vpc/latest/userguide/subnet-sizing.html)
- [Subnets for your VPC](https://docs.aws.amazon.com/vpc/latest/userguide/configure-subnets.html)
- [net/netip: AddrFrom4, As4 and PrefixFrom](https://pkg.go.dev/net/netip#AddrFrom4)
