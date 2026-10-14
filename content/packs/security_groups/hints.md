## Task 1: Check a security group

Loop over the rules and return `true` as soon as one matches. A rule
matches when its CIDR contains the source and its protocol and ports cover
the packet; a `-1` rule covers any protocol and port.

### Closer

`r.CIDR.Contains(p.Source)` checks the source. For the rest, a helper keeps
task 2 short:

```go
func portMatches(ruleProtocol string, from, to int, protocol string, port int) bool {
	if ruleProtocol == AllProtocols {
		return true
	}
	// same protocol, and port between from and to
}
```

### Answer

```go
func portMatches(ruleProtocol string, from, to int, protocol string, port int) bool {
	if ruleProtocol == AllProtocols {
		return true
	}
	return ruleProtocol == protocol && from <= port && port <= to
}

func SGAllows(rules []SGRule, p Packet) bool {
	for _, r := range rules {
		if r.CIDR.Contains(p.Source) && portMatches(r.Protocol, r.FromPort, r.ToPort, p.Protocol, p.Port) {
			return true
		}
	}
	return false
}
```

## Task 2: Check a network ACL

Sort a copy of the rules by `Number`, then walk it: the first matching rule
returns its `Allow`. If the loop ends, the `*` rule denies.

### Closer

`sorted := slices.Clone(rules)`, then
`slices.SortFunc(sorted, func(a, b NACLRule) int { return cmp.Compare(a.Number, b.Number) })`.
Add `"cmp"` and `"slices"` to the imports. Match each rule with
`r.CIDR.Contains(remote)` and your helper from task 1.

### Answer

```go
sorted := slices.Clone(rules)
slices.SortFunc(sorted, func(a, b NACLRule) int {
	return cmp.Compare(a.Number, b.Number)
})
for _, r := range sorted {
	if r.CIDR.Contains(remote) && portMatches(r.Protocol, r.FromPort, r.ToPort, protocol, port) {
		return r.Allow
	}
}
return false
```

## Task 3: Follow a packet both ways

Three `if` checks in the table's order, each returning its message, then
`"allowed"`. Only the reply check uses a different port.

### Closer

Inbound is `NACLAllows(acl.Inbound, p.Source, p.Protocol, p.Port)` and the
security group is `SGAllows(sg, p)`. For the reply, call `NACLAllows` on
`acl.Outbound` with `p.SourcePort` as the port.

### Answer

```go
if !NACLAllows(acl.Inbound, p.Source, p.Protocol, p.Port) {
	return "blocked by network ACL (inbound)"
}
if !SGAllows(sg, p) {
	return "blocked by security group"
}
if !NACLAllows(acl.Outbound, p.Source, p.Protocol, p.SourcePort) {
	return "blocked by network ACL (outbound reply)"
}
return "allowed"
```
