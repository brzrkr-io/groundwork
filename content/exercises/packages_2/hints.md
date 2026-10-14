## Task 1: Alias the two region imports

Both packages are called `region`, so Go can't tell them apart. You can give each import its own local name.

### Closer

An alias goes in front of the import path: `awsregion "groundwork.local/packages_2/aws/region"`. Then use `awsregion.Valid` and `gcpregion.Valid` in `Provider`.

### Answer

```go
import (
	awsregion "groundwork.local/packages_2/aws/region"
	gcpregion "groundwork.local/packages_2/gcp/region"
)

func Provider(name string) string {
	switch {
	case awsregion.Valid(name):
		return "aws"
	case gcpregion.Valid(name):
		return "gcp"
	default:
		return "unknown"
	}
}
```

## Task 2: Add Partition to the AWS package

The function goes in `aws/region/region.go`. Its name starts with a capital letter so other packages can call it. It checks how the name starts.

### Closer

Use `strings.HasPrefix(name, "us-gov-")` and `strings.HasPrefix(name, "cn-")` in a `switch { ... }`, and fall back to `"aws"`.

### Answer

```go
func Partition(name string) string {
	switch {
	case strings.HasPrefix(name, "us-gov-"):
		return "aws-us-gov"
	case strings.HasPrefix(name, "cn-"):
		return "aws-cn"
	default:
		return "aws"
	}
}
```

## Task 3: Finish Describe

Ask `Provider` which cloud the name belongs to, and build a different sentence for each answer.

### Closer

`switch Provider(name) { case "aws": ... }`. The AWS sentence adds `awsregion.Partition(name)` on the end.

### Answer

```go
switch Provider(name) {
case "aws":
	return name + " is AWS, partition " + awsregion.Partition(name)
case "gcp":
	return name + " is GCP"
default:
	return name + " is not a known region"
}
```
