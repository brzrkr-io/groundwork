## Task 1: Match a wildcard pattern

Make every character of the pattern literal first, then give `*` and `?`
their wildcard meaning back. Anchor the expression so the whole value has to
match.

### Closer

`regexp.QuoteMeta(pattern)` turns `*` into `\*` and `?` into `\?`. Replace
those two with `.*` and `.` using `strings.ReplaceAll` (raw strings make it
readable: `` `\*` ``), then compile `"^" + expr + "$"` and call
`MatchString(value)`.

### Answer

```go
expr := regexp.QuoteMeta(pattern)
expr = strings.ReplaceAll(expr, `\*`, ".*")
expr = strings.ReplaceAll(expr, `\?`, ".")
return regexp.MustCompile("^" + expr + "$").MatchString(value)
```

## Task 2: Check whether a statement applies

Two separate questions, both of which must be yes: does any action pattern
match, and does any resource pattern match? For actions, make both sides the
same case before you compare.

### Closer

Loop over `s.Action` and call `Match(strings.ToLower(p), strings.ToLower(action))`;
remember whether one matched in a `bool`. If none did, return `false`. Then
loop over `s.Resource` with `Match(p, resource)` and return `true` on the
first hit.

### Answer

```go
actionMatches := false
for _, p := range s.Action {
	if Match(strings.ToLower(p), strings.ToLower(action)) {
		actionMatches = true
		break
	}
}
if !actionMatches {
	return false
}
for _, p := range s.Resource {
	if Match(p, resource) {
		return true
	}
}
return false
```

## Task 3: Decide a request

Two nested loops: policies, then their statements. Skip statements that don't
apply. A `Deny` ends the search at once; an `Allow` only means "allowed so
far", because a later `Deny` still wins.

### Closer

Keep `allowed := false`. Inside the inner loop: `if !Applies(s, action,
resource) { continue }`, then return `ExplicitDeny` when `s.Effect ==
"Deny"`, or set `allowed = true` when it is `"Allow"`. After both loops,
`allowed` picks between `Allow` and `ImplicitDeny`.

### Answer

```go
allowed := false
for _, p := range policies {
	for _, s := range p.Statement {
		if !Applies(s, action, resource) {
			continue
		}
		if s.Effect == "Deny" {
			return ExplicitDeny
		}
		if s.Effect == "Allow" {
			allowed = true
		}
	}
}
if allowed {
	return Allow
}
return ImplicitDeny
```
