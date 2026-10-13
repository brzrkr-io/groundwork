## Task 1: loop over the required keys

Don't loop over the instance's tags; `have` already holds them. Loop over
`RequiredTags` instead, and look each key up in `have`. A key that is absent
reads as `""`, just like a key with an empty value, so one check covers both:

```go
var missing []string
for _, key := range RequiredTags {
	if have[key] == "" {
		missing = append(missing, key)
	}
}
return missing
```

## Task 2: skip compliant instances, then sort

Inside the inner loop, call `MissingTags(inst.Tags)` first. If the result is
empty, `continue`. Otherwise append
`Finding{InstanceID: aws.ToString(inst.InstanceId), Missing: missing}`.

After the paginator loop:

```go
slices.SortFunc(findings, func(a, b Finding) int {
	return strings.Compare(a.InstanceID, b.InstanceID)
})
```

## Task 3: only keys with a default

Use the two-value map lookup to see whether a default exists:

```go
for _, key := range f.Missing {
	if value, ok := defaults[key]; ok {
		tags = append(tags, types.Tag{Key: aws.String(key), Value: aws.String(value)})
	}
}
if len(tags) == 0 {
	continue
}
```

Then add `DryRun: aws.Bool(dryRun),` to the `CreateTagsInput` literal.
