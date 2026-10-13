# Paginators: walk every page

A log bucket can hold millions of objects, and a big AWS account can run
thousands of instances. AWS never sends all of that in one reply. It sends
*pages*, and your code has to ask for each one. Forget that, and your cleanup
script quietly skips almost everything.

## Learn

### Why AWS returns pages

When you list an S3 bucket with `ListObjectsV2`, S3 sends back at most 1,000
keys per call. If more keys remain, the reply says so and includes a
*continuation token*: a bookmark saying "carry on from here". To get the next
page, you call `ListObjectsV2` again with that token. When the reply has no
token, you have reached the last page. This is called *pagination*.

Every AWS list or describe call works like this. Only the field names differ:

| Operation | "More pages?" | Token in the reply | Token in the next request |
|---|---|---|---|
| S3 `ListObjectsV2` | `IsTruncated` is true | `NextContinuationToken` | `ContinuationToken` |
| EC2 `DescribeInstances` | the token is not empty | `NextToken` | `NextToken` |

Tokens are *opaque*: you never read or build one yourself, you only pass it
back.

### Paginators do the bookkeeping

Juggling tokens by hand is fiddly, so the SDK gives you a *paginator* for
each operation that pages. You create it once with the same input you would
pass to the call. Then you loop:

```go
p := s3.NewListObjectsV2Paginator(api, &s3.ListObjectsV2Input{
	Bucket: aws.String("app-logs"),
	Prefix: aws.String("web/"),
})

for p.HasMorePages() {
	page, err := p.NextPage(ctx) // one ListObjectsV2 call
	if err != nil {
		return nil, fmt.Errorf("list s3://app-logs/web/: %w", err)
	}
	for _, obj := range page.Contents {
		fmt.Println(aws.ToString(obj.Key))
	}
}
```

- `HasMorePages()` is true before the first page, and afterwards for as long
  as the last reply carried a token.
- `NextPage(ctx)` makes one API call and returns that page: a normal
  `*s3.ListObjectsV2Output`. The paginator stores the new token for the next
  call.
- `page.Contents` holds the objects on this page. Each object's `Key` is a
  `*string`, so read it with `aws.ToString`, as in *SDK clients: small interfaces and fakes*.

EC2 works the same way. Each page holds `Reservations`, and each reservation
holds `Instances`:

```go
p := ec2.NewDescribeInstancesPaginator(api, &ec2.DescribeInstancesInput{})
// ... same loop; then for each r in page.Reservations,
// for each inst in r.Instances, read aws.ToString(inst.InstanceId).
```

### Interfaces that come with the SDK

In *SDK clients: small interfaces and fakes* you wrote a small interface listing just the
methods you call. The SDK already ships one for every operation that pages:

```go
type ListObjectsV2APIClient interface {
	ListObjectsV2(context.Context, *s3.ListObjectsV2Input, ...func(*s3.Options)) (*s3.ListObjectsV2Output, error)
}
```

`s3.ListObjectsV2APIClient` and `ec2.DescribeInstancesAPIClient` are what the
paginators accept. `*s3.Client` and `*ec2.Client` satisfy them, so in
production you pass the real client. The tests pass a fake that hands out
small pages. The fakes are in `fakes_test.go`, if you want to see how they
make tokens.

### Page size is not a total

`MaxKeys` on the input sets the *page size*: "send me at most N keys per
page". It does not stop the loop. The paginator keeps fetching pages until
the bucket runs out. To stop early, `return` (or `break`) out of the loop
yourself. Every page you skip is an API call you don't pay for.

### The context reaches every call

`NextPage(ctx)` hands `ctx` to the API call it makes. If `ctx` is cancelled
(Ctrl-C, a deadline, a shutdown, as in *Context and cancellation*), the next
call fails with an error that wraps `context.Canceled`. Return it wrapped
with `%w`, so callers can still check `errors.Is(err, context.Canceled)`.

## Your task

Your team's ops tool lists log files in S3 and instances in EC2. It works in
the dev account, which is small, but misses things in production.

### 1. Read every page

`AllKeys` already creates a paginator, but it reads only the first page. Put
the page-reading code inside a `for p.HasMorePages() { ... }` loop, so it
returns the key of every object whose key starts with `prefix`, in the order
S3 returns them. On error, return `nil` and the error wrapped as
`list s3://<bucket>/<prefix>: <err>` (the starter already does this).

```go
AllKeys(ctx, api, "app-logs", "web/") // the bucket sends 2 keys per page
// => ["web/2026-10-09.log" "web/2026-10-10.log" "web/2026-10-11.log"], nil
```

### 2. Stop once you have enough

`FirstKeys` should return the first `limit` keys of the bucket. It sets
`MaxKeys`, but as you read above, that only sets the page size, so right now
it fetches every page. Return `keys` as soon as it holds `limit` keys, so no
further pages are requested. If the bucket has fewer keys than `limit`,
return all of them. `limit` is always at least 1.

```go
FirstKeys(ctx, api, "deploy-artifacts", 3) // the bucket sends 2 keys per page
// => the first 3 keys, after only 2 ListObjectsV2 calls
```

### 3. List every instance

Write `InstanceIDs`. Use `ec2.NewDescribeInstancesPaginator` with an empty
`&ec2.DescribeInstancesInput{}` and return the `InstanceId` of every instance
in every reservation on every page, in order. Pass `ctx` to `NextPage`.

On error, return `nil` and the error wrapped with `%w` as
`describe instances: <err>`. The tests cancel the context after the first
page, so you should get:

```go
InstanceIDs(ctx, api) // ctx is cancelled after page 1
// => nil, error "describe instances: context canceled"
//    errors.Is(err, context.Canceled) == true, and no third page is requested
```

## Learn more

- [Using paginators (AWS SDK for Go v2)](https://docs.aws.amazon.com/sdk-for-go/v2/developer-guide/using.html#using-operation-paginators)
- [ListObjectsV2 (Amazon S3 API reference)](https://docs.aws.amazon.com/AmazonS3/latest/API/API_ListObjectsV2.html)
- [s3.ListObjectsV2Paginator on pkg.go.dev](https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/service/s3#ListObjectsV2Paginator)
- [ec2.DescribeInstancesPaginator on pkg.go.dev](https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/service/ec2#DescribeInstancesPaginator)
