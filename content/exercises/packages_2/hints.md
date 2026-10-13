## Task 1: name before path

An alias goes in front of the import path, inside the import block:

```go
import (
	awsregion "gopherlings.local/packages_2/aws/region"
	gcpregion "gopherlings.local/packages_2/gcp/region"
)
```

Then `region.Valid` in `Provider` has to become `awsregion.Valid` or
`gcpregion.Valid`.

## Task 2: an exported function in another folder

The function goes in `aws/region/region.go`, starts with a capital letter, and
checks prefixes with `strings.HasPrefix`. Check `us-gov-` and `cn-` before
falling back to `"aws"`.

## Task 3: switch on the provider

```go
switch Provider(name) {
case "aws":
	return name + " is AWS, partition " + awsregion.Partition(name)
case "gcp":
	return name + " is GCP"
}
return name + " is not a known region"
```
