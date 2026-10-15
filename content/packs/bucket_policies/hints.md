## Task 1: Spot a public statement

Rule out the easy cases first: not an Allow, or no `"*"` in `Principal.AWS`.
Then walk every condition key, under every operator, and look for one of
`RestrictingKeys`.

### Closer

`slices.Contains(s.Principal.AWS, "*")` checks the principal. Three loops:
`for _, keys := range s.Condition`, then `for key := range keys`, then over
`RestrictingKeys`, comparing with `strings.EqualFold(key, restricting)`.

### Answer

```go
if s.Effect != "Allow" || !slices.Contains(s.Principal.AWS, "*") {
	return false
}
for _, keys := range s.Condition {
	for key := range keys {
		for _, restricting := range RestrictingKeys {
			if strings.EqualFold(key, restricting) {
				return false
			}
		}
	}
}
return true
```

## Task 2: Decide a bucket's exposure

Compute two booleans, "public policy" and "public ACL", then turn the table
into code, top row first.

### Closer

Loop over `b.Policy.Statement` calling `PublicStatement`. For the ACL, two
`slices.Contains` calls joined with `||`. Then:

```go
switch {
case publicPolicy && !b.Settings.RestrictPublicBuckets, publicACL && ...:
	return "public"
case ...:
	return "blocked"
default:
	return "private"
}
```

A comma between `case` conditions means "either one".

### Answer

```go
publicPolicy := false
for _, s := range b.Policy.Statement {
	if PublicStatement(s) {
		publicPolicy = true
	}
}
publicACL := slices.Contains(b.ACLGrantees, AllUsers) || slices.Contains(b.ACLGrantees, AuthenticatedUsers)

switch {
case publicPolicy && !b.Settings.RestrictPublicBuckets, publicACL && !b.Settings.IgnorePublicAcls:
	return "public"
case publicPolicy || publicACL:
	return "blocked"
default:
	return "private"
}
```

## Task 3: Check that HTTPS is required

Loop over the statements and return `true` at the first one that passes all
four checks. Each check is one expression.

### Closer

`s.Condition["Bool"]["aws:SecureTransport"]` is a `StringList` (empty when
missing), so `slices.Contains(..., "false")` works on it. Join the four
checks with `&&`. Join the two action checks with `||`, in parentheses.

### Answer

```go
for _, s := range p.Statement {
	if s.Effect == "Deny" &&
		slices.Contains(s.Principal.AWS, "*") &&
		(slices.Contains(s.Action, "s3:*") || slices.Contains(s.Action, "*")) &&
		slices.Contains(s.Condition["Bool"]["aws:SecureTransport"], "false") {
		return true
	}
}
return false
```
