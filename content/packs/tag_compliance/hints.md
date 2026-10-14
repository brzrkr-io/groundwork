## Task 1: Find missing tags

Don't loop over the instance's tags; `have` already holds them. Loop over
`RequiredTags` instead, and ask what `have` says about each key.

### Closer

A key that is absent reads as `""` from a map, just like a key with an empty
value, so one check covers both: `if have[key] == "" { ... }`. Collect those
keys in a `var missing []string`.

### Answer

```go
var missing []string
for _, key := range RequiredTags {
	if have[key] == "" {
		missing = append(missing, key)
	}
}
return missing
```

## Task 2: Audit every instance

Inside the inner loop, ask `MissingTags` about the instance before you append
anything. After the paginator loop, sort the findings.

### Closer

If `len(missing) == 0`, `continue`. Otherwise fill in both fields of
`Finding`. Sort with `slices.SortFunc(findings, func(a, b Finding) int { ... })`
and `strings.Compare`, and add `"slices"` and `"strings"` to the imports.

### Answer

```go
missing := MissingTags(inst.Tags)
if len(missing) == 0 {
	continue
}
findings = append(findings, Finding{
	InstanceID: aws.ToString(inst.InstanceId),
	Missing:    missing,
})

// after the paginator loop:
slices.SortFunc(findings, func(a, b Finding) int {
	return strings.Compare(a.InstanceID, b.InstanceID)
})
```

## Task 3: Build the fixes

Loop over `f.Missing` and only use keys that have an entry in `defaults`. If
no tag was added, skip the finding.

### Closer

The two-value map lookup tells you whether a default exists:
`if value, ok := defaults[key]; ok { ... }`. Then add
`DryRun: aws.Bool(dryRun),` to the `CreateTagsInput` literal.

### Answer

```go
for _, key := range f.Missing {
	if value, ok := defaults[key]; ok {
		tags = append(tags, types.Tag{Key: aws.String(key), Value: aws.String(value)})
	}
}
if len(tags) == 0 {
	continue
}

fixes = append(fixes, &ec2.CreateTagsInput{
	Resources: []string{f.InstanceID},
	Tags:      tags,
	DryRun:    aws.Bool(dryRun),
})
```
