# Security groups and network ACLs: will this packet get through?

"I can't reach the server" is one of the most common cloud tickets. In AWS,
two firewalls sit between a client and your server, and they follow
different rules. Here you write a tool that answers the ticket: allowed, or
which firewall stopped it.

**New here:** security groups, network ACLs, stateful vs stateless firewalls, ephemeral ports.

## Learn

### Two firewalls on the way

Traffic to a server in a VPC passes two firewalls. The *network ACL*
(access control list) guards the whole subnet. The *security group* guards
the server's network card. The reply travels back out the same way:

```text
client 198.51.100.7:50432 ──► network ACL (inbound) ──► security group ──► server :443
client 198.51.100.7:50432 ◄── network ACL (outbound) ◄── (reply: no check) ◄── server
```

| | Security group | Network ACL |
|---|---|---|
| Attached to | a server's network interface | a subnet |
| Rules | allow only | numbered, each ALLOW or DENY |
| Which rule wins | any matching rule allows | the lowest-numbered matching rule decides |
| Nothing matches | denied | denied (the final `*` rule) |
| Replies | allowed automatically (*stateful*) | need their own outbound rule (*stateless*) |

### Security groups

A security group is a list of allow rules. Each rule says: this protocol,
these ports, from this CIDR. If any rule matches, the traffic is in. A new
security group has no inbound rules, so it lets nothing in.

| Protocol | Ports | Source | Means |
|---|---|---|---|
| `tcp` | 443–443 | `0.0.0.0/0` | HTTPS from anywhere |
| `tcp` | 22–22 | `203.0.113.0/24` | SSH from the office only |
| `-1` | (all) | `10.0.5.0/24` | everything from the monitoring subnet |

The protocol `-1` means "all protocols"; its ports don't matter, a `-1` rule
covers every port. In this exercise protocols are written `"tcp"` and
`"udp"`, lowercase, as in AWS's security group API. Both port numbers of a
range are included.

### Network ACLs

A network ACL is a numbered list. AWS checks the rules from the lowest
number up, and the **first** rule that matches decides, ALLOW or DENY.
Every ACL ends with a rule numbered `*` that denies everything.

| Rule # | Protocol | Ports | Source | Action |
|---|---|---|---|---|
| 90 | all | all | `198.51.100.0/24` | DENY |
| 100 | tcp | 443 | `0.0.0.0/0` | ALLOW |
| `*` | all | all | `0.0.0.0/0` | DENY |

A client in `198.51.100.0/24` is denied by rule 90, even though rule 100
would allow it. The order in which the rules are *stored* doesn't matter,
only their numbers. A VPC's default network ACL allows everything; a new
custom one has only the `*` rule, so it blocks everything.

### Replies and ephemeral ports

A client connects *from* a random high port, called an *ephemeral port*:
`198.51.100.7:50432` connects to your server's port `443`. The reply goes
from your port 443 back **to port 50432**.

- The security group is *stateful*: it remembers the connection and lets
  the reply out without any rule.
- The network ACL is *stateless*: the reply is just more traffic, checked
  against the outbound rules with the client as destination and **the
  client's port** as the port. So outbound rules usually allow TCP
  1024–65535.

The classic mistake: an outbound ACL that only allows port 443. Requests
arrive, but every reply to port 50432 is dropped, and clients time out.

### Recap: sorting a copy

As in *Sorting*, `slices.SortFunc` sorts in place, so sort a copy when the
slice belongs to the caller:

```go
mine := slices.Clone(deploys)
slices.SortFunc(mine, func(a, b Deploy) int {
	return cmp.Compare(a.Started, b.Started)
})
```

The tests use `netip.MustParseAddr` and `netip.MustParsePrefix`, which
return the value directly and panic on bad input, as in
*VPC planning: subnets per availability zone*.

### Try it (optional)

These commands only read. Needs an AWS account and credentials; never
needed for the tests.

```bash
aws ec2 describe-security-groups --group-ids sg-0123456789abcdef0   # a security group and its rules
aws ec2 describe-security-group-rules --filters Name=group-id,Values=sg-0123456789abcdef0  # one line per rule
aws ec2 describe-network-acls --filters Name=association.subnet-id,Values=subnet-0123456789abcdef0  # the ACL on a subnet, with rule numbers
```

Replace the IDs with your own from `aws ec2 describe-instances`.

## Your task

A teammate can't reach a web server. Write the tool that follows a packet
through both firewalls. The types are in `exercise.go`; tests build them
like `SGRule{"tcp", 443, 443, netip.MustParsePrefix("0.0.0.0/0")}`.

### 1. Check a security group

**Goal:** `SGAllows(rules, p)` returns `true` if any rule lets `p` in.

A rule matches when all of these hold:

- its `CIDR` contains `p.Source`;
- its `Protocol` is `AllProtocols` (`"-1"`), **or** it equals `p.Protocol`
  and `FromPort <= p.Port <= ToPort`.

No rules (or `nil`) allows nothing.

```go
rules := []SGRule{
	{"tcp", 443, 443, netip.MustParsePrefix("0.0.0.0/0")},
	{"tcp", 22, 22, netip.MustParsePrefix("203.0.113.0/24")},
	{"-1", 0, 0, netip.MustParsePrefix("10.0.5.0/24")},
}
SGAllows(rules, Packet{netip.MustParseAddr("198.51.100.7"), 50432, "tcp", 443}) // => true
SGAllows(rules, Packet{netip.MustParseAddr("198.51.100.7"), 50433, "tcp", 22})  // => false (SSH only from the office)
SGAllows(rules, Packet{netip.MustParseAddr("198.51.100.7"), 50434, "udp", 443}) // => false (wrong protocol)
SGAllows(rules, Packet{netip.MustParseAddr("10.0.5.9"), 40003, "udp", 161})    // => true (-1 covers all)
```

**Think it through:**
- Which `Packet` fields does a security group rule look at, and which one does it ignore?
- The network ACL in task 2 matches traffic the same way. A small helper for the protocol-and-port part saves writing it twice.

### 2. Check a network ACL

**Goal:** `NACLAllows(rules, remote, protocol, port)` returns the `Allow` of the lowest-numbered rule that matches, or `false` when none does.

- A rule matches like a security group rule, with `remote` in place of the
  source and `port` in place of `p.Port`.
- `rules` can come in any order; go by `Number`, lowest first.
- Don't change the order of the caller's slice (a test checks).

```go
rules := []NACLRule{
	{200, "tcp", 443, 443, netip.MustParsePrefix("0.0.0.0/0"), true},
	{100, "-1", 0, 0, netip.MustParsePrefix("198.51.100.0/24"), false},
	{300, "tcp", 1024, 65535, netip.MustParsePrefix("0.0.0.0/0"), true},
}
NACLAllows(rules, netip.MustParseAddr("198.51.100.7"), "tcp", 443) // => false (100 denies first)
NACLAllows(rules, netip.MustParseAddr("192.0.2.44"), "tcp", 443)   // => true (200)
NACLAllows(rules, netip.MustParseAddr("192.0.2.44"), "tcp", 50432) // => true (300)
NACLAllows(rules, netip.MustParseAddr("192.0.2.44"), "tcp", 22)    // => false (only "*" matches)
```

**Think it through:**
- What has to happen to the rules before the "first match" loop, and to which slice?
- Once a rule matches, what do you return, and why can you stop looking?

### 3. Follow a packet both ways

**Goal:** `Reach(acl, sg, p)` returns exactly one of four strings, checking in packet order.

| Step | Check | If it fails, return |
|---|---|---|
| 1 | `acl.Inbound` allows `p.Source`, `p.Protocol`, `p.Port` | `"blocked by network ACL (inbound)"` |
| 2 | the security group `sg` allows `p` | `"blocked by security group"` |
| 3 | `acl.Outbound` allows the reply: to `p.Source`, same protocol, port **`p.SourcePort`** | `"blocked by network ACL (outbound reply)"` |

If all three pass, return `"allowed"`. The security group needs no reply
check: it is stateful.

```go
allowAll := []NACLRule{{100, "-1", 0, 0, netip.MustParsePrefix("0.0.0.0/0"), true}}
only443 := []NACLRule{{100, "tcp", 443, 443, netip.MustParsePrefix("0.0.0.0/0"), true}}
web := []SGRule{{"tcp", 443, 443, netip.MustParsePrefix("0.0.0.0/0")}}
client := Packet{netip.MustParseAddr("198.51.100.7"), 50432, "tcp", 443}

Reach(NACL{allowAll, allowAll}, web, client) // => "allowed"
Reach(NACL{allowAll, only443}, web, client)  // => "blocked by network ACL (outbound reply)"
Reach(NACL{}, web, client)                   // => "blocked by network ACL (inbound)"
```

**Think it through:**
- Which of your two functions answers each step, and which arguments change for the reply?
- Why does the order of the checks matter when two firewalls would both block?

## Learn more

- [Control traffic to your AWS resources using security groups](https://docs.aws.amazon.com/vpc/latest/userguide/vpc-security-groups.html)
- [Control subnet traffic with network access control lists](https://docs.aws.amazon.com/vpc/latest/userguide/vpc-network-acls.html)
- [Network ACL rules](https://docs.aws.amazon.com/vpc/latest/userguide/nacl-rules.html)
- [Custom network ACLs for your VPC (example rules with ephemeral ports)](https://docs.aws.amazon.com/vpc/latest/userguide/custom-network-acl.html)
