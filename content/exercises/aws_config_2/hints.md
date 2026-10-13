## Task 1: copy, change, append

```go
for _, region := range regions {
	c := cfg.Copy()
	c.Region = region
	out = append(out, c)
}
```

## Task 2: the condition is backwards

The override should be applied when `region` is set, not when `cfg.Region`
is empty. Either change the `if` to check `region != ""`, or pass
`config.WithRegion(region)` to `LoadDefaultConfig` (only when `region` is not
`""`), the way *AWS SDK configuration: regions, profiles and credentials*
built its options slice.

## Task 3: build the line from safe parts

Never put `creds.SecretAccessKey` or `creds.SessionToken` into the string.
Turn the token into a yes/no first:

```go
temporary := "no"
if creds.SessionToken != "" {
	temporary = "yes"
}
return fmt.Sprintf("region=%s key=%s temporary=%s", cfg.Region, mask(creds.AccessKeyID), temporary), nil
```
