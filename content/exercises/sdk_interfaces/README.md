# SDK clients: small interfaces and fakes

Real AWS code talks to real AWS, and you can't run that in a unit test. In
this exercise you call S3 and EC2 through small interfaces, so your tests can
swap in a fake and run offline in milliseconds.

## Learn

### Clients and their methods

The AWS SDK for Go gives you one *client* per service. You build it from the
`aws.Config` you loaded in *AWS SDK configuration: regions, profiles and credentials*:

```go
cfg, err := config.LoadDefaultConfig(ctx)
// ...
s3Client := s3.NewFromConfig(cfg)   // *s3.Client
ec2Client := ec2.NewFromConfig(cfg) // *ec2.Client
```

Every API call AWS offers is a method on the client, and every method has the
same shape:

```go
func (c *Client) ListBuckets(ctx context.Context, params *ListBucketsInput, optFns ...func(*Options)) (*ListBucketsOutput, error)
```

| Part | What it is |
|---|---|
| `ctx` | A `context.Context` that can cancel the call, as in *Context and cancellation*. |
| `params` | A pointer to an `XxxInput` struct with the request fields. |
| `optFns` | Optional per-call settings. You almost never pass any. |
| result | A pointer to an `XxxOutput` struct, and an `error`. |

### Depend on a small interface

`*s3.Client` has over a hundred methods, and each one sends a real HTTP
request. Your code only calls one or two of them. So, as in *Fakes and
dependency injection*, declare an interface with just the methods you call,
and accept that interface instead of the client:

```go
// BucketLister is the one S3 method this code needs.
type BucketLister interface {
	ListBuckets(ctx context.Context, params *s3.ListBucketsInput, optFns ...func(*s3.Options)) (*s3.ListBucketsOutput, error)
}
```

Copy the method line exactly from the SDK (pkg.go.dev, or "go to definition"
in your editor). `*s3.Client` has a method with that exact signature, so it
satisfies `BucketLister` automatically. In production you pass the real
client, and in tests you pass a fake.

To be sure the real client still fits, add a *compile-time assertion*:

```go
var _ BucketLister = (*s3.Client)(nil)
```

It says "a `*s3.Client` must be usable as a `BucketLister`". The `_` throws
the value away. If the signatures ever differ, the build fails on this line
and tells you which method doesn't match, instead of failing somewhere far
away in `main`.

### Pointers everywhere

Almost every field in an SDK struct is a pointer, such as `Name *string`,
because AWS needs to tell "not set" (`nil`) apart from "set to empty". The
`aws` package (`github.com/aws/aws-sdk-go-v2/aws`) has helpers to go both
ways:

```go
in := &s3.ListBucketsInput{Prefix: aws.String("prod-")} // string -> *string
name := aws.ToString(bucket.Name)                       // *string -> string, nil gives ""
```

There are versions for other types too, such as `aws.Int32` and
`aws.ToInt32`. Always read pointer fields with `aws.ToString` and friends.
Writing `*bucket.Name` panics if the field is `nil`.

### S3 buckets

*S3* stores files, which it calls *objects*, in named containers called
*buckets*, such as `acme-app-logs`. `ListBuckets` returns the buckets in your
account:

```go
out, err := api.ListBuckets(ctx, &s3.ListBucketsInput{})
if err != nil {
	return nil, fmt.Errorf("list buckets: %w", err)
}
for _, b := range out.Buckets { // b is an s3types.Bucket
	fmt.Println(aws.ToString(b.Name))
}
```

### EC2 instances

An *EC2 instance* is a virtual machine that AWS runs for you: a server with an
ID like `i-0abc123`, which you start, stop and terminate through the API.
`DescribeInstances` lists them. The answer is nested, because AWS groups
instances that were launched together into a *reservation*:

```go
out, err := api.DescribeInstances(ctx, &ec2.DescribeInstancesInput{})
// ...
for _, r := range out.Reservations {   // ec2types.Reservation
	for _, inst := range r.Instances { // ec2types.Instance
		id := aws.ToString(inst.InstanceId)
		if inst.State != nil && inst.State.Name == ec2types.InstanceStateNameRunning {
			fmt.Println(id, "is running")
		}
	}
}
```

`inst.State` is a pointer, so check it for `nil` first. `State.Name` is an
`ec2types.InstanceStateName`, a string type with constants such as
`InstanceStateNameRunning` (`"running"`) and `InstanceStateNameStopped`
(`"stopped"`). Instances also carry *tags*, key/value labels such as
`Env=prod`, in `inst.Tags` (a `[]ec2types.Tag` with `Key` and `Value`, both
`*string`).

The S3 and EC2 packages both have a `types` package, so give them short names
when you import them:

```go
s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
```

## Your task

Your team's inventory tool reports which buckets exist and which servers are
running. Make it testable.

When you finish, the import block of `exercise.go` holds `context`, `fmt`,
`slices`, `github.com/aws/aws-sdk-go-v2/aws`, the `ec2` and `s3` packages, and
the two `types` packages named `ec2types` and `s3types`. Add each one when a
task needs it.

### 1. List bucket names

Implement `BucketNames(ctx, api)`. Call `api.ListBuckets` with
`&s3.ListBucketsInput{}`, collect each bucket's name with `aws.ToString`, and
return the names sorted with `slices.Sort`. If the call fails, return `nil`
and the error wrapped as `list buckets: %w`.

```go
BucketNames(ctx, api) // AWS has "web-assets", "app-logs", "backups"
// => []string{"app-logs", "backups", "web-assets"}

BucketNames(ctx, api) // AWS answers with error "AccessDenied"
// => nil, error "list buckets: AccessDenied"
```

The tests pass in a fake that works like the one you build in task 2.

### 2. Finish the fake

`FakeBucketLister` is a fake S3 client for tests. Finish its `ListBuckets`
method:

- If `f.Err` is set, return `nil` and `f.Err`.
- Otherwise return a `*s3.ListBucketsOutput` with one `s3types.Bucket` per
  name in `f.Names`, in the same order, with `Name` set via `aws.String`.
  With no names, return an output with no buckets (not `nil`).

```go
f := &FakeBucketLister{Names: []string{"web-assets", "app-logs"}}
out, err := f.ListBuckets(ctx, &s3.ListBucketsInput{})
// aws.ToString(out.Buckets[0].Name) => "web-assets"
// aws.ToString(out.Buckets[1].Name) => "app-logs"
// err => nil
```

The test also checks `var _ BucketLister = (*FakeBucketLister)(nil)`, the same
kind of assertion as for the real client.

### 3. Describe your own interface

`InstanceDescriber` is empty. Add the `DescribeInstances` method of
`*ec2.Client` to it. Copy the signature exactly: it takes
`*ec2.DescribeInstancesInput` and `...func(*ec2.Options)`, and returns
`*ec2.DescribeInstancesOutput` and an `error`. The line
`var _ InstanceDescriber = (*ec2.Client)(nil)` is already in the file. If it
goes red after your change, your signature doesn't match the real one.

Then implement `RunningInstanceIDs(ctx, api)`. Call `api.DescribeInstances`
with `&ec2.DescribeInstancesInput{}`, go through every reservation and every
instance, and collect the `InstanceId` of each instance whose state is
`ec2types.InstanceStateNameRunning`. Skip instances with a `nil` `State`.
Return the IDs sorted. If the call fails, return `nil` and the error wrapped
as `describe instances: %w`.

```go
RunningInstanceIDs(ctx, api)
// reservation 1: i-0b2 running, i-0c3 stopped
// reservation 2: i-0a1 running
// => []string{"i-0a1", "i-0b2"}
```

Read only the first page of results. Accounts with many instances get their
answer split into pages, which you handle in *Paginators: walk every page*.

## Learn more

- [AWS SDK for Go v2: Unit testing with the SDK](https://docs.aws.amazon.com/sdk-for-go/v2/developer-guide/unit-testing.html)
- [pkg.go.dev: service/s3 Client.ListBuckets](https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/service/s3#Client.ListBuckets)
- [pkg.go.dev: service/ec2 Client.DescribeInstances](https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/service/ec2#Client.DescribeInstances)
- [Amazon EC2: Instance lifecycle](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ec2-instance-lifecycle.html)
