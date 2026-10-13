## Task 1: call, loop, sort

`api.ListBuckets(ctx, &s3.ListBucketsInput{})` gives you `out` and `err`.
Check `err` first. Then loop over `out.Buckets`, append
`aws.ToString(b.Name)` to a slice, and call `slices.Sort` on it.

## Task 2: build the output

Start with `if f.Err != nil { return nil, f.Err }`. Then, for each name, append
a bucket to `out.Buckets`:

```go
out.Buckets = append(out.Buckets, s3types.Bucket{Name: aws.String(name)})
```

## Task 3: the interface

Copy the method line from `*ec2.Client`, without the receiver and the body:

```go
type InstanceDescriber interface {
	DescribeInstances(ctx context.Context, params *ec2.DescribeInstancesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error)
}
```

## Task 3: two loops

```go
for _, r := range out.Reservations {
	for _, inst := range r.Instances {
		if inst.State != nil && inst.State.Name == ec2types.InstanceStateNameRunning {
			ids = append(ids, aws.ToString(inst.InstanceId))
		}
	}
}
```

Sort `ids` before you return them, and wrap the error from
`DescribeInstances` as `describe instances: %w`.
