## Task 1: One config per region

Loop over `regions`. Each time round, make a copy of `cfg`, change the copy's
region, and append the copy to `out`.

### Closer

`cfg.Copy()` gives you a new `aws.Config`. Store it in a variable, set its
`Region` field, then `append` it:

```go
for _, region := range regions {
	c := cfg.Copy()
	// set c's Region, then append c to out
}
```

### Answer

```go
for _, region := range regions {
	c := cfg.Copy()
	c.Region = region
	out = append(out, c)
}
```

Changing `c.Region` leaves `cfg` alone, and every copy shares the same
credentials provider.

## Task 2: Fix the ignored region override

Look at the `if`: it only uses `region` when `cfg.Region` is empty. The
override should apply whenever `region` itself is not `""`.

### Closer

Instead of patching `cfg` afterwards, pass `config.WithRegion(region)` to
`LoadDefaultConfig`, but only when `region != ""`. Build the options in a
slice first:

```go
opts := []func(*config.LoadOptions) error{config.WithSharedConfigProfile(profile)}
if region != "" {
	// append config.WithRegion(region)
}
```

### Answer

```go
opts := []func(*config.LoadOptions) error{config.WithSharedConfigProfile(profile)}
if region != "" {
	opts = append(opts, config.WithRegion(region))
}
return config.LoadDefaultConfig(ctx, opts...)
```

`WithRegion` beats `AWS_REGION` and the profile's `region =` line, and an
unknown profile still makes `LoadDefaultConfig` return an error.

## Task 3: Stop leaking credentials

Never put `creds.SecretAccessKey` or `creds.SessionToken` into the string.
You only need to know whether a session token exists, as `yes` or `no`.

### Closer

Turn the token into a word first, then format the line with three `%s`:

```go
temporary := "no"
if creds.SessionToken != "" {
	temporary = ...
}
```

### Answer

```go
temporary := "no"
if creds.SessionToken != "" {
	temporary = "yes"
}
return fmt.Sprintf("region=%s key=%s temporary=%s", cfg.Region, mask(creds.AccessKeyID), temporary), nil
```

The line now only holds the region, the masked key ID and a yes/no.
