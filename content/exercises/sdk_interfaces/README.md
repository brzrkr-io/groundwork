# SDK clients: small interfaces and fakes

Real AWS code talks to real AWS, and you can't run that in a unit test. In
this exercise you call S3 and EC2 through small interfaces, so your tests can
swap in a fake and run offline in milliseconds.

**New here:** SDK clients, `Input`/`Output` structs, compile-time interface assertions, `aws.String`/`aws.ToString`, S3 buckets, EC2 instances.

## Learn

### Clients and their methods

The AWS SDK for Go gives you one *client* per service. You build it from the
`aws.Config` you loaded in *AWS SDK configuration: regions, profiles and credentials*:

```go
s3Client := s3.NewFromConfig(cfg)   // *s3.Client
ec2Client := ec2.NewFromConfig(cfg) // *ec2.Client
```

Every API call is a method on the client, and every method has the same
shape. Here is one you won't need in the tasks:

```go
func (c *Client) GetBucketLocation(ctx context.Context, params *GetBucketLocationInput, optFns ...func(*Options)) (*GetBucketLocationOutput, error)
```

| Part | What it is |
|---|---|
| `(c *Client)` | the *receiver*: this is a method on `*Client` |
| `ctx` | a `context.Context` that can cancel the call, as in *Context and cancellation* |
| `params` | a pointer to an `XxxInput` struct with the request fields |
| `optFns` | optional per-call settings; you almost never pass any |
| results | a pointer to an `XxxOutput` struct, and an `error` |

### Depend on a small interface

`*s3.Client` has over a hundred methods, each sending a real HTTP request.
As in *Fakes and dependency injection*, declare an interface with only the
methods you call, and accept that interface:

```go
// LocationGetter is the one S3 method this code needs.
type LocationGetter interface {
	GetBucketLocation(ctx context.Context, params *s3.GetBucketLocationInput, optFns ...func(*s3.Options)) (*s3.GetBucketLocationOutput, error)
}
```

An interface method line is the method's line **without** `func`, without the
receiver and without the body. Outside package `s3`, the type names need the
`s3.` prefix.

> **Why is it written like this?** Go requires the method name, parameter
> types and result types to match the client exactly, or the client won't
> satisfy the interface. The parameter names (`ctx`, `params`, `optFns`) are
> convention: Go ignores them in interfaces, but copying them keeps it
> readable. The interface name is your choice; `-er` names are convention.

Copy the line from pkg.go.dev or your editor's "go to definition".

### Compile-time assertions

To be sure the real client still fits, add:

```go
var _ LocationGetter = (*s3.Client)(nil)
```

It says "a `*s3.Client` must be usable as a `LocationGetter`". `_` throws the
value away; `(*s3.Client)(nil)` is a typed `nil` pointer. If the signatures
ever differ, the build fails on this line and names the mismatched method.

### Pointers everywhere

Almost every SDK field is a pointer, such as `Name *string`, so AWS can tell
"not set" (`nil`) from "set to empty". The `aws` package
(`github.com/aws/aws-sdk-go-v2/aws`) converts both ways:

```go
in := &s3.ListObjectsV2Input{Bucket: aws.String("acme-reports")} // string -> *string
owner := aws.ToString(out.Owner.DisplayName)                     // *string -> string, nil gives ""
```

There are versions for other types too, such as `aws.Int32` and
`aws.ToInt32`. Writing `*field` panics when the field is `nil`, so prefer the
helpers.

### S3 buckets

*S3* stores files (*objects*) in named containers called *buckets*, such as
`acme-app-logs`. `ListBuckets` returns the buckets in your account in
`out.Buckets`, a `[]s3types.Bucket`. Each bucket has `Name *string` and
`CreationDate *time.Time`.

```go
for _, b := range out.Buckets {
	fmt.Println(aws.ToTime(b.CreationDate))
}
```

### EC2 instances

An *EC2 instance* is a virtual machine with an ID like `i-0abc123`.
`DescribeInstances` lists them, nested: AWS groups instances launched together
into a *reservation*. `out.Reservations` is a `[]ec2types.Reservation`, and
each reservation's `Instances` is a `[]ec2types.Instance`.

Each instance has `InstanceId *string` and `State *ec2types.InstanceState`.
`State.Name` is an `ec2types.InstanceStateName`, a string type with constants:

| Constant | Value |
|---|---|
| `ec2types.InstanceStateNamePending` | `"pending"` |
| `ec2types.InstanceStateNameRunning` | `"running"` |
| `ec2types.InstanceStateNameStopped` | `"stopped"` |
| `ec2types.InstanceStateNameTerminated` | `"terminated"` |

`State` is a pointer, so check `nil` before reading `State.Name`. Go's `&&`
stops at the first `false`, which makes this safe:

```go
if inst.State != nil && inst.State.Name == ec2types.InstanceStateNameStopped {
	stopped++
}
```

Both services have a `types` package, so rename them when you import them:

```go
s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
```

> **Why is it written like this?** Go requires different names for two
> imports that would both be called `types`. The names `s3types` and
> `ec2types` are convention in AWS code.

## Your task

Your team's inventory tool reports which buckets exist and which servers are
running. Make it testable.

When you finish, the import block holds `context`, `fmt`, `slices`,
`github.com/aws/aws-sdk-go-v2/aws`, the `ec2` and `s3` packages, and the two
`types` packages named `ec2types` and `s3types`. Add each one when a task
needs it.

### 1. List bucket names

**Goal:** `BucketNames(ctx, api)` returns the names of all buckets, sorted.

Call `ListBuckets` with an empty `&s3.ListBucketsInput{}`. Read each name with
`aws.ToString` and sort with `slices.Sort`. With no buckets, return `nil`
(an empty, unset slice). If the call fails, return `nil` and the error wrapped
with `%w` as `list buckets: <err>`.

```go
BucketNames(ctx, api) // AWS has "web-assets", "app-logs", "backups"
// => []string{"app-logs", "backups", "web-assets"}, nil

BucketNames(ctx, api) // AWS answers with error "AccessDenied"
// => nil, error "list buckets: AccessDenied"
```

**Think it through:**
- Which field of the output holds the buckets, and which field of a bucket holds its name?
- What do you check right after the call, before touching the output?

### 2. Finish the fake

**Goal:** `FakeBucketLister.ListBuckets` returns its error if one is set,
otherwise one bucket per name in `f.Names`.

- If `f.Err` is set, return `nil` and `f.Err`.
- Otherwise return a `*s3.ListBucketsOutput` with one `s3types.Bucket` per
  name, in the same order, with `Name` set. With no names, return an output
  with no buckets (not `nil`).

```go
f := &FakeBucketLister{Names: []string{"web-assets", "app-logs"}}
out, err := f.ListBuckets(ctx, &s3.ListBucketsInput{})
// aws.ToString(out.Buckets[0].Name) => "web-assets"
// aws.ToString(out.Buckets[1].Name) => "app-logs"
// err => nil
```

The test also checks `var _ BucketLister = (*FakeBucketLister)(nil)`.

**Think it through:**
- `Name` is a `*string` but `f.Names` holds plain strings: which helper converts?
- The starter already creates `out`. What do you add to it in the loop?

### 3. Describe your own interface

**Goal:** add the `DescribeInstances` method to `InstanceDescriber`, then make
`RunningInstanceIDs(ctx, api)` return the IDs of all running instances, sorted.

The method must match `*ec2.Client` exactly: it takes `ctx`,
`*ec2.DescribeInstancesInput` and `...func(*ec2.Options)`, and returns
`*ec2.DescribeInstancesOutput` and an `error`. The line
`var _ InstanceDescriber = (*ec2.Client)(nil)` is already in the file; if it
goes red, your line doesn't match.

`RunningInstanceIDs` calls it with an empty `&ec2.DescribeInstancesInput{}`
and goes through every reservation and every instance. Keep an instance only
if its `State` is not `nil` and its name is `running`; pending, stopped,
stopping and terminated instances are skipped. Return the IDs sorted, or
`nil` if none are running. If the call fails, return `nil` and the error
wrapped with `%w` as `describe instances: <err>`.

```go
RunningInstanceIDs(ctx, api)
// reservation 1: i-0b2 running, i-0c3 stopped
// reservation 2: i-0a1 running
// => []string{"i-0a1", "i-0b2"}, nil
```

Read only the first page of results. Pages come in *Paginators: walk every page*.

**Think it through:**
- Compare the `BucketLister` interface with the `ListBuckets` method on the fake: which parts did the interface drop?
- How many `for` loops do you need to reach each instance?
- Which check must come first so you never read through a `nil` `State`?

## Learn more

- [AWS SDK for Go v2: Unit testing with the SDK](https://docs.aws.amazon.com/sdk-for-go/v2/developer-guide/unit-testing.html)
- [pkg.go.dev: service/s3 Client.ListBuckets](https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/service/s3#Client.ListBuckets)
- [pkg.go.dev: service/ec2 Client.DescribeInstances](https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/service/ec2#Client.DescribeInstances)
- [Amazon EC2: Instance lifecycle](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ec2-instance-lifecycle.html)
