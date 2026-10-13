## Task 1: List bucket names

Call `api.ListBuckets` and check the error first. Then loop over the buckets
in the output, collect their names in a slice, and sort it.

### Closer

`api.ListBuckets(ctx, &s3.ListBucketsInput{})` returns `out` and `err`. Each
`b` in `out.Buckets` has a `Name *string`, so read it with
`aws.ToString(b.Name)`. Wrap the error with
`fmt.Errorf("list buckets: %w", err)`. Add the imports `fmt`, `slices` and
`github.com/aws/aws-sdk-go-v2/aws`.

### Answer

```go
out, err := api.ListBuckets(ctx, &s3.ListBucketsInput{})
if err != nil {
	return nil, fmt.Errorf("list buckets: %w", err)
}
var names []string
for _, b := range out.Buckets {
	names = append(names, aws.ToString(b.Name))
}
slices.Sort(names)
return names, nil
```

## Task 2: Finish the fake

Two cases: if the fake has an error, hand it back. Otherwise build one bucket
per name in `f.Names` and add it to `out.Buckets`.

### Closer

Start with `if f.Err != nil { return nil, f.Err }`. A bucket is an
`s3types.Bucket` whose `Name` is a `*string`, so use `aws.String(name)`. Add
the import `s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"`.

### Answer

```go
if f.Err != nil {
	return nil, f.Err
}
out := &s3.ListBucketsOutput{}
for _, name := range f.Names {
	out.Buckets = append(out.Buckets, s3types.Bucket{Name: aws.String(name)})
}
return out, nil
```

## Task 3: Describe your own interface

First put the `DescribeInstances` method line into `InstanceDescriber`,
copied from `*ec2.Client`. Then call it and walk two nested loops:
reservations, then instances.

### Closer

An interface method is the method line without `func`, the receiver and the
body. In the loop, keep an instance only if
`inst.State != nil && inst.State.Name == ec2types.InstanceStateNameRunning`.
Add the import `ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"`.

```go
for _, r := range out.Reservations {
	for _, inst := range r.Instances {
		// check State, then append aws.ToString(inst.InstanceId)
	}
}
```

### Answer

```go
type InstanceDescriber interface {
	DescribeInstances(ctx context.Context, params *ec2.DescribeInstancesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error)
}
```

```go
out, err := api.DescribeInstances(ctx, &ec2.DescribeInstancesInput{})
if err != nil {
	return nil, fmt.Errorf("describe instances: %w", err)
}
var ids []string
for _, r := range out.Reservations {
	for _, inst := range r.Instances {
		if inst.State != nil && inst.State.Name == ec2types.InstanceStateNameRunning {
			ids = append(ids, aws.ToString(inst.InstanceId))
		}
	}
}
slices.Sort(ids)
return ids, nil
```

Checking `inst.State != nil` first means `inst.State.Name` is never read
through a `nil` pointer.
