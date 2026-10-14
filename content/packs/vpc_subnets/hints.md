## Task 1: Carve out one subnet

Add the two missing error checks after the IPv4 one, each returning
`netip.Prefix{}` and an error. Then block number `index` starts `index`
blocks after the VPC's first address, and one block holds
`1 << (32 - bits)` addresses.

### Closer

The number of blocks is `1 << (bits - vpc.Bits())`. To move `start`,
convert first and shift: `start += uint32(index) << (32 - bits)`. Errors
come from `fmt.Errorf("...")`, like the IPv4 check's.

### Answer

```go
// after the IPv4 check:
if bits < vpc.Bits() || bits > 28 {
	return netip.Prefix{}, fmt.Errorf("/%d doesn't fit in %s: want /%d to /28", bits, vpc, vpc.Bits())
}
blocks := 1 << (bits - vpc.Bits())
if index < 0 || index >= blocks {
	return netip.Prefix{}, fmt.Errorf("index %d out of range: %s has %d blocks of /%d", index, vpc, blocks, bits)
}
start := toUint32(vpc.Masked().Addr())
start += uint32(index) << (32 - bits)
return netip.PrefixFrom(fromUint32(start), bits), nil
```

## Task 2: Plan subnets for every zone

Parse the VPC string first. Then each subnet is one `SubnetCIDR` call: the
public subnet for `zones[i]` uses index `i`, the private one uses
`len(zones) + i`.

### Closer

Loop over the two kinds, and inside that over the zones:

```go
for k, kind := range []string{"public", "private"} {
	for i, zone := range zones {
		cidr, err := SubnetCIDR(prefix, bits, k*len(zones)+i)
		// check err, then append a Subnet
	}
}
```

`Usable` is `1<<(32-bits) - Reserved`.

### Answer

```go
prefix, err := netip.ParsePrefix(vpc)
if err != nil {
	return nil, err
}
var plan []Subnet
for k, kind := range []string{"public", "private"} {
	for i, zone := range zones {
		cidr, err := SubnetCIDR(prefix, bits, k*len(zones)+i)
		if err != nil {
			return nil, err
		}
		plan = append(plan, Subnet{
			Name:   kind + "-" + zone,
			Zone:   zone,
			CIDR:   cidr,
			Public: kind == "public",
			Usable: 1<<(32-bits) - Reserved,
		})
	}
}
return plan, nil
```

## Task 3: Check a hand-written plan

For each subnet, ask three questions in the table's order and append a
message for each "yes". For the overlap question, compare the subnet only
with the ones after it.

### Closer

`s.CIDR.Bits() > 28` means too small. Outside the VPC is
`!prefix.Contains(s.CIDR.Addr()) || s.CIDR.Bits() < prefix.Bits()`. For the
pairs, `for i, s := range subnets` and inside it
`for _, other := range subnets[i+1:]` with `s.CIDR.Overlaps(other.CIDR)`.
Build each message with `fmt.Sprintf` and `%s`.

### Answer

```go
prefix, err := netip.ParsePrefix(vpc)
if err != nil {
	return []string{"invalid VPC CIDR: " + vpc}
}
var problems []string
for i, s := range subnets {
	if s.CIDR.Bits() > 28 {
		problems = append(problems, fmt.Sprintf("%s: %s is smaller than /28", s.Name, s.CIDR))
	}
	if !prefix.Contains(s.CIDR.Addr()) || s.CIDR.Bits() < prefix.Bits() {
		problems = append(problems, fmt.Sprintf("%s: %s is not inside %s", s.Name, s.CIDR, vpc))
	}
	for _, other := range subnets[i+1:] {
		if s.CIDR.Overlaps(other.CIDR) {
			problems = append(problems, fmt.Sprintf("%s (%s) overlaps %s (%s)", s.Name, s.CIDR, other.Name, other.CIDR))
		}
	}
}
return problems
```
