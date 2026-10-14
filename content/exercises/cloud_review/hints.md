## Task 1: Find admin ports open to the internet

Loop over the groups and their `Inbound` rules. Skip any rule whose `CIDR`
isn't the whole internet, then check the ports `22` and `3389` one by one.

### Closer

A rule opens `port` when `r.Protocol == "-1"`, or when `r.Protocol == "tcp"`
and `r.FromPort <= port && port <= r.ToPort`. Build the problem with
`fmt.Sprintf("port %d open to %s", port, r.CIDR)` and the resource with
`g.ID + " (" + g.Name + ")"`.

### Answer

```go
var findings []Finding
for _, g := range groups {
	for _, r := range g.Inbound {
		if r.CIDR != "0.0.0.0/0" && r.CIDR != "::/0" {
			continue
		}
		for _, port := range []int{22, 3389} {
			all := r.Protocol == "-1"
			inRange := r.Protocol == "tcp" && r.FromPort <= port && port <= r.ToPort
			if all || inRange {
				findings = append(findings, Finding{
					Severity: "high",
					Resource: g.ID + " (" + g.Name + ")",
					Problem:  fmt.Sprintf("port %d open to %s", port, r.CIDR),
				})
			}
		}
	}
}
return findings
```

## Task 2: Find over-broad policies

Each policy's statements are in `p.Document.Statement`. Skip statements that
aren't `Allow` or don't have `"*"` among their resources, then look at each
action.

### Closer

`slices.Contains(s.Resource, "*")` checks the resources, and
`strings.HasSuffix(action, ":*")` spots a service wildcard. Check
`action == "*"` first, because `"*"` doesn't end in `":*"`. Import `slices`
and `strings`.

### Answer

```go
var findings []Finding
for _, p := range policies {
	for _, s := range p.Document.Statement {
		if s.Effect != "Allow" || !slices.Contains(s.Resource, "*") {
			continue
		}
		for _, action := range s.Action {
			f := Finding{Resource: "policy " + p.Name}
			switch {
			case action == "*":
				f.Severity, f.Problem = "high", "allows every action on every resource"
			case strings.HasSuffix(action, ":*"):
				f.Severity, f.Problem = "medium", "allows "+action+" on every resource"
			default:
				continue
			}
			findings = append(findings, f)
		}
	}
}
return findings
```

## Task 3: Find public buckets

Skip a bucket as soon as you see `BlockPublicAccess` is on. For the others,
check each statement of `b.Policy` against the four conditions in the README.

### Closer

`len(s.Condition) == 0` is true for a statement without a condition (a nil
map has length 0). `slices.Contains(s.Principal.AWS, "*")` checks the
principal, and `strings.Join(s.Action, ", ")` lists the actions.

### Answer

```go
var findings []Finding
for _, b := range buckets {
	if b.BlockPublicAccess {
		continue
	}
	for _, s := range b.Policy.Statement {
		if s.Effect == "Allow" && slices.Contains(s.Principal.AWS, "*") && len(s.Condition) == 0 {
			findings = append(findings, Finding{
				Severity: "high",
				Resource: "bucket " + b.Name,
				Problem:  "anyone can " + strings.Join(s.Action, ", "),
			})
		}
	}
}
return findings
```

## Task 4: Find stale access keys

Skip inactive keys. Work out the age in days first, then make up to two
findings, the age one before the never-used one.

### Closer

`now.Sub(k.Created).Hours() / 24` is the age in days as a `float64`;
`int(...)` drops the fraction, which rounds down. `k.LastUsed.IsZero()`
reports a key that was never used.

### Answer

```go
var findings []Finding
for _, k := range keys {
	if k.Status != "Active" {
		continue
	}
	resource := "user " + k.User
	if days := int(now.Sub(k.Created).Hours() / 24); days > 90 {
		findings = append(findings, Finding{
			Severity: "medium",
			Resource: resource,
			Problem:  fmt.Sprintf("access key %s is %d days old", k.ID, days),
		})
	}
	if k.LastUsed.IsZero() {
		findings = append(findings, Finding{
			Severity: "low",
			Resource: resource,
			Problem:  "access key " + k.ID + " has never been used",
		})
	}
}
return findings
```

## Task 5: Write the report

Copy the slice, sort the copy, then write one line per finding and count
each severity as you go. Severities need a number to sort by.

### Closer

A map like `map[string]int{"high": 0, "medium": 1, "low": 2}` turns a
severity into a rank. In `slices.SortFunc`, compare ranks with
`cmp.Compare`, and only when that returns `0` compare `Resource`, then
`Problem`. A `strings.Builder` with `fmt.Fprintf` collects the lines. Import
`cmp`.

### Answer

```go
sorted := slices.Clone(findings)
slices.SortFunc(sorted, func(a, b Finding) int {
	if c := cmp.Compare(severityRank[a.Severity], severityRank[b.Severity]); c != 0 {
		return c
	}
	if c := cmp.Compare(a.Resource, b.Resource); c != 0 {
		return c
	}
	return cmp.Compare(a.Problem, b.Problem)
})
var sb strings.Builder
counts := map[string]int{}
for _, f := range sorted {
	fmt.Fprintf(&sb, "[%s] %s: %s\n", f.Severity, f.Resource, f.Problem)
	counts[f.Severity]++
}
fmt.Fprintf(&sb, "%d high, %d medium, %d low\n", counts["high"], counts["medium"], counts["low"])
return sb.String()
```

with `severityRank` declared next to `Report`:

```go
var severityRank = map[string]int{"high": 0, "medium": 1, "low": 2}
```
