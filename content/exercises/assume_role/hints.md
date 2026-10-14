## Task 1: Spot temporary credentials

Two checks, both on fields of `c`: how the key ID starts, and whether the
session token is empty.

### Closer

`strings.HasPrefix(s, prefix)` returns a `bool`. Join the two checks with `&&`.

### Answer

```go
return strings.HasPrefix(c.AccessKeyID, "ASIA") && c.SessionToken != ""
```

## Task 2: Refresh before expiry

Work out how long is left from `now` until `c.Expiration`, then compare that
duration with five minutes. Expired credentials have a negative time left.

### Closer

`c.Expiration.Sub(now)` is a `time.Duration`; `5*time.Minute` is five minutes.
"Less than" is `<`, so exactly five minutes gives `false`.

### Answer

```go
return c.Expiration.Sub(now) < 5*time.Minute
```

## Task 3: Check who may assume a role

Return `false` early if the effect isn't `"Allow"` or no action matches. Then
try the three ways the principal can trust the caller. Only ARNs have an
account, so the root check needs the caller split on `:`.

### Closer

For the action, loop over `s.Action` and test
`Match(strings.ToLower(a), "sts:assumerole")`. For the principal,
`slices.Contains(s.Principal.AWS, caller)` and friends. For the account,
`parts := strings.SplitN(caller, ":", 6)`: a full ARN gives 6 parts and the
account is `parts[4]`.

### Answer

```go
func TrustsCaller(s TrustStatement, caller string) bool {
	if s.Effect != "Allow" || !allowsAction(s.Action, "sts:AssumeRole") {
		return false
	}
	aws := s.Principal.AWS
	if slices.Contains(aws, caller) || slices.Contains(s.Principal.Service, caller) || slices.Contains(aws, "*") {
		return true
	}
	parts := strings.SplitN(caller, ":", 6)
	return len(parts) == 6 && slices.Contains(aws, "arn:aws:iam::"+parts[4]+":root")
}

// allowsAction reports whether one of the action patterns matches action,
// ignoring case.
func allowsAction(patterns StringList, action string) bool {
	for _, p := range patterns {
		if Match(strings.ToLower(p), strings.ToLower(action)) {
			return true
		}
	}
	return false
}
```

The `len(parts) == 6` check keeps `parts[4]` from panicking on a service name.

## Task 4: Check which workflows may deploy

Start like task 3 (effect and action, now `sts:AssumeRoleWithWebIdentity`),
then check `Federated`. After that, every condition key must find its claim
and at least one matching value; one failure is enough to say no.

### Closer

Two nested loops: `for operator, keys := range s.Condition` and
`for key, values := range keys`. Look up `claim, ok := claims[key]`. A helper
`anyValueMatches(operator, values, claim)` can loop over the values with a
`switch operator` inside: `case "StringEquals":` compares with `==`,
`case "StringLike":` calls `Match(v, claim)`.

### Answer

```go
func TrustsWorkflow(s TrustStatement, provider string, claims map[string]string) bool {
	if s.Effect != "Allow" || !allowsAction(s.Action, "sts:AssumeRoleWithWebIdentity") {
		return false
	}
	if !slices.Contains(s.Principal.Federated, provider) {
		return false
	}
	for operator, keys := range s.Condition {
		for key, values := range keys {
			claim, ok := claims[key]
			if !ok || !anyValueMatches(operator, values, claim) {
				return false
			}
		}
	}
	return true
}

// anyValueMatches reports whether claim satisfies operator for at least one
// of values. Unknown operators never match.
func anyValueMatches(operator string, values StringList, claim string) bool {
	for _, v := range values {
		switch operator {
		case "StringEquals":
			if claim == v {
				return true
			}
		case "StringLike":
			if Match(v, claim) {
				return true
			}
		}
	}
	return false
}
```

An unknown operator falls through the `switch`, so the helper returns `false`.
