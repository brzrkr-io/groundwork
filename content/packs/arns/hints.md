## Task 1: Parse an ARN

An ARN has six parts, but the resource may contain more colons. Split with a
limit so they stay in the last part, then check the length before you look at
any part.

### Closer

`parts := strings.SplitN(s, ":", 6)` gives at most six parts. Return an error
when `len(parts) != 6 || parts[0] != "arn"`. Then fill in the struct from
`parts[1]` to `parts[5]` and check the three fields that must not be empty.

### Answer

```go
parts := strings.SplitN(s, ":", 6)
if len(parts) != 6 || parts[0] != "arn" {
	return ARN{}, fmt.Errorf("%q is not an ARN", s)
}
a := ARN{Partition: parts[1], Service: parts[2], Region: parts[3], Account: parts[4], Resource: parts[5]}
if a.Partition == "" || a.Service == "" || a.Resource == "" {
	return ARN{}, fmt.Errorf("%q is not an ARN: empty partition, service or resource", s)
}
return a, nil
```

## Task 2: Tell the scope

Read the table from the top: a region makes it regional, whatever the
account. Only when there's no region does the account matter.

### Closer

A `switch` with no value after it runs the first `case` that is true:
`case a.Region != "":`, then `case a.Account != "":`, then `default:`.

### Answer

```go
switch {
case a.Region != "":
	return "regional"
case a.Account != "":
	return "account"
default:
	return "global"
}
```

## Task 3: Find a zone's region

Split the job in two: the last character, and everything before it. Deal with
the empty string first, because indexing it panics.

### Closer

`zone[len(zone)-1]` is the last byte and `zone[:len(zone)-1]` the rest. Bytes
compare like runes: `letter < 'a' || letter > 'z'` means "not a lowercase
letter". `strings.Count(region, "-")` counts the dashes. Check the count
before you look at the region's last byte, so a one-letter zone can't panic.

### Answer

```go
if zone == "" {
	return "", fmt.Errorf("empty zone name %q", zone)
}
letter := zone[len(zone)-1]
region := zone[:len(zone)-1]
if letter < 'a' || letter > 'z' {
	return "", fmt.Errorf("%q must end in a letter a-z", zone)
}
if strings.Count(region, "-") < 2 || region[len(region)-1] < '0' || region[len(region)-1] > '9' {
	return "", fmt.Errorf("%q is not a region", region)
}
return region, nil
```

## Task 4: Find resources in other accounts

Loop over the strings, parse each one with `Parse`, and stop at the first
error. Keep the original string when the account is someone else's.

### Closer

Inside the loop: `a, err := Parse(s)`, return `nil, err` on an error, then
`if a.Account != "" && a.Account != account { foreign = append(foreign, s) }`.
Start with `var foreign []string`.

### Answer

```go
var foreign []string
for _, s := range arns {
	a, err := Parse(s)
	if err != nil {
		return nil, err
	}
	if a.Account != "" && a.Account != account {
		foreign = append(foreign, s)
	}
}
return foreign, nil
```
